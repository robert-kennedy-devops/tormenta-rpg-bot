// Package skills provides validation and auditing utilities for the skill system.
// It operates on the rpg.SkillNode definitions and detects common design
// problems: duplicate IDs, mechanically-identical skills, missing roles,
// unbalanced scaling and orphaned prerequisites.
package skills

import (
	"fmt"
	"math"
	"strings"

	"github.com/tormenta-bot/internal/rpg"
)

// ─── Audit report ─────────────────────────────────────────────────────────────

// Issue severity levels.
type Severity string

const (
	SevError   Severity = "ERROR"
	SevWarning Severity = "WARNING"
	SevInfo    Severity = "INFO"
)

// Issue describes a single finding from the validator.
type Issue struct {
	Severity Severity
	SkillID  string
	Code     string
	Message  string
}

func (i Issue) String() string {
	return fmt.Sprintf("[%s] %s (%s): %s", i.Severity, i.SkillID, i.Code, i.Message)
}

// Report is the full output of ValidateSkillTrees.
type Report struct {
	Issues     []Issue
	TotalSkills int
	Classes    map[string]int // class → skill count
}

// HasErrors returns true if the report contains any ERROR-level issues.
func (r *Report) HasErrors() bool {
	for _, i := range r.Issues {
		if i.Severity == SevError {
			return true
		}
	}
	return false
}

// Summary returns a human-readable one-liner.
func (r *Report) Summary() string {
	errs, warns := 0, 0
	for _, i := range r.Issues {
		switch i.Severity {
		case SevError:
			errs++
		case SevWarning:
			warns++
		}
	}
	return fmt.Sprintf("Skills: %d total | Errors: %d | Warnings: %d", r.TotalSkills, errs, warns)
}

// ─── Validator ────────────────────────────────────────────────────────────────

// ValidateSkillTrees runs all checks against the global rpg.Trees registry.
// It is safe to call at startup (e.g. in main) so problems surface immediately.
func ValidateSkillTrees() Report {
	report := Report{Classes: map[string]int{}}

	// Flatten all nodes.
	all := allNodes()
	report.TotalSkills = len(all)
	for _, n := range all {
		report.Classes[n.Class]++
	}

	report.Issues = append(report.Issues, checkDuplicateIDs(all)...)
	report.Issues = append(report.Issues, checkMissingRoles(all)...)
	report.Issues = append(report.Issues, checkDuplicateMechanics(all)...)
	report.Issues = append(report.Issues, checkOrphanedRequirements(all)...)
	report.Issues = append(report.Issues, checkBalanceAnomalies(all)...)
	report.Issues = append(report.Issues, checkPassiveMarked(all)...)

	return report
}

// ─── Individual checks ────────────────────────────────────────────────────────

// checkDuplicateIDs detects two or more nodes sharing the same ID string.
func checkDuplicateIDs(nodes []rpg.SkillNode) []Issue {
	seen := map[string]string{} // id → first class
	var issues []Issue
	for _, n := range nodes {
		if prev, ok := seen[n.ID]; ok {
			issues = append(issues, Issue{
				Severity: SevError,
				SkillID:  n.ID,
				Code:     "DUPLICATE_ID",
				Message:  fmt.Sprintf("ID já existe em classe '%s'", prev),
			})
		} else {
			seen[n.ID] = n.Class
		}
	}
	return issues
}

// checkMissingRoles warns about nodes with an empty Role field.
func checkMissingRoles(nodes []rpg.SkillNode) []Issue {
	var issues []Issue
	for _, n := range nodes {
		if n.Role == "" {
			issues = append(issues, Issue{
				Severity: SevWarning,
				SkillID:  n.ID,
				Code:     "MISSING_ROLE",
				Message:  fmt.Sprintf("'%s' não tem SkillRole definido", n.Name),
			})
		}
	}
	return issues
}

// mechanicKey builds a canonical key that summarises the mechanical signature
// of a skill: effect types + elements + status kinds (sorted, joined).
// Two skills with the same key are mechanically identical.
func mechanicKey(n rpg.SkillNode) string {
	if len(n.Effects) == 0 {
		if n.IsPassive {
			return "passive:" + n.ID // passives are always unique by definition
		}
		return "noeffect"
	}
	parts := make([]string, 0, len(n.Effects))
	for _, e := range n.Effects {
		parts = append(parts, strings.Join([]string{
			e.TypeName,
			e.Element,
			e.StatusKind,
		}, ":"))
	}
	// sort so order doesn't matter
	for i := 1; i < len(parts); i++ {
		for j := i; j > 0 && parts[j] < parts[j-1]; j-- {
			parts[j], parts[j-1] = parts[j-1], parts[j]
		}
	}
	return strings.Join(parts, "|")
}

// checkDuplicateMechanics detects skills that are mechanically identical.
// Two skills are identical if they share the same mechanicKey AND the same Role.
func checkDuplicateMechanics(nodes []rpg.SkillNode) []Issue {
	type sig struct {
		mkey string
		role rpg.SkillRole
	}
	seen := map[sig]string{} // sig → first skill ID
	var issues []Issue
	for _, n := range nodes {
		if n.IsPassive || n.Role == rpg.RolePassive {
			continue
		}
		s := sig{mkey: mechanicKey(n), role: n.Role}
		if s.mkey == "noeffect" {
			continue
		}
		if prev, ok := seen[s]; ok {
			issues = append(issues, Issue{
				Severity: SevWarning,
				SkillID:  n.ID,
				Code:     "DUPLICATE_MECHANIC",
				Message:  fmt.Sprintf("'%s' tem mecânica idêntica a '%s' (role=%s, chave=%s)", n.Name, prev, n.Role, s.mkey),
			})
		} else {
			seen[s] = n.ID
		}
	}
	return issues
}

// checkOrphanedRequirements detects prerequisites that point to non-existent IDs.
func checkOrphanedRequirements(nodes []rpg.SkillNode) []Issue {
	ids := make(map[string]bool, len(nodes))
	for _, n := range nodes {
		ids[n.ID] = true
	}
	var issues []Issue
	for _, n := range nodes {
		for _, req := range n.Requires {
			if !ids[req] {
				issues = append(issues, Issue{
					Severity: SevError,
					SkillID:  n.ID,
					Code:     "ORPHAN_REQUIREMENT",
					Message:  fmt.Sprintf("'%s' requer '%s' que não existe", n.Name, req),
				})
			}
		}
	}
	return issues
}

// checkBalanceAnomalies flags skills whose BasePower is an outlier compared to
// the mean + 2 standard deviations of other skills at the same Tier.
func checkBalanceAnomalies(nodes []rpg.SkillNode) []Issue {
	// group BasePower by tier
	byTier := map[int][]float64{}
	for _, n := range nodes {
		for _, e := range n.Effects {
			if e.TypeName == "damage" && e.BasePower > 0 {
				byTier[n.Tier] = append(byTier[n.Tier], float64(e.BasePower))
			}
		}
	}

	threshold := map[int]float64{} // tier → mean + 2σ
	for tier, vals := range byTier {
		if len(vals) < 3 {
			continue
		}
		mean := 0.0
		for _, v := range vals {
			mean += v
		}
		mean /= float64(len(vals))
		variance := 0.0
		for _, v := range vals {
			d := v - mean
			variance += d * d
		}
		variance /= float64(len(vals))
		threshold[tier] = mean + 2*math.Sqrt(variance)
	}

	var issues []Issue
	for _, n := range nodes {
		limit, ok := threshold[n.Tier]
		if !ok {
			continue
		}
		for _, e := range n.Effects {
			if e.TypeName == "damage" && float64(e.BasePower) > limit {
				issues = append(issues, Issue{
					Severity: SevWarning,
					SkillID:  n.ID,
					Code:     "BALANCE_ANOMALY",
					Message: fmt.Sprintf(
						"'%s' BasePower=%d está acima de mean+2σ=%.0f para tier %d",
						n.Name, e.BasePower, limit, n.Tier),
				})
			}
		}
	}
	return issues
}

// checkPassiveMarked warns if a node has IsPassive=true but Role != RolePassive
// (or vice-versa), which indicates an inconsistency.
func checkPassiveMarked(nodes []rpg.SkillNode) []Issue {
	var issues []Issue
	for _, n := range nodes {
		if n.IsPassive && n.Role != "" && n.Role != rpg.RolePassive {
			issues = append(issues, Issue{
				Severity: SevWarning,
				SkillID:  n.ID,
				Code:     "PASSIVE_ROLE_MISMATCH",
				Message: fmt.Sprintf(
					"'%s' IsPassive=true mas Role=%s (esperado RolePassive)", n.Name, n.Role),
			})
		}
		if !n.IsPassive && n.Role == rpg.RolePassive {
			issues = append(issues, Issue{
				Severity: SevWarning,
				SkillID:  n.ID,
				Code:     "PASSIVE_ROLE_MISMATCH",
				Message: fmt.Sprintf(
					"'%s' Role=PASSIVE mas IsPassive=false", n.Name),
			})
		}
	}
	return issues
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

// allNodes flattens all SkillNodes from rpg.Trees into a single slice.
func allNodes() []rpg.SkillNode {
	var out []rpg.SkillNode
	for _, tree := range rpg.Trees {
		for _, branch := range tree.Branches {
			out = append(out, branch.Nodes...)
		}
	}
	return out
}
