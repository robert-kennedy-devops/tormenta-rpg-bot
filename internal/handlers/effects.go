package handlers

import "fmt"

// TempEffects guarda buffs/debuffs temporários por entidade em combate.
type TempEffects struct {
	CABonus      int
	CABonusTurns int

	CAPenalty      int
	CAPenaltyTurns int

	AtkPenalty      int
	AtkPenaltyTurns int

	DmgReductionPct      int
	DmgReductionPctTurns int

	ForceOneDamageTurns int

	OutgoingDmgPct  int
	OutgoingDmgHits int

	ForceCritHits int
	CritMin       int

	EnemyDotDmg   int
	EnemyDotTurns int

	SkipEnemyAttackTurns int
	SkipOwnTurnTurns     int

	// Named monster status for synergy checks (AppliesStatus / RequiresStatus)
	MonsterStatus      string
	MonsterStatusTurns int
}

func (e *TempEffects) EffectiveCABonus() int {
	if e.CABonusTurns > 0 {
		return e.CABonus
	}
	return 0
}

func (e *TempEffects) EffectiveCAPenalty() int {
	if e.CAPenaltyTurns > 0 {
		return e.CAPenalty
	}
	return 0
}

func (e *TempEffects) EffectiveAtkPenalty() int {
	if e.AtkPenaltyTurns > 0 {
		return e.AtkPenalty
	}
	return 0
}

func (e *TempEffects) ApplyIncomingDamage(dmg int) int {
	if dmg <= 0 {
		return dmg
	}
	if e.ForceOneDamageTurns > 0 {
		return 1
	}
	if e.DmgReductionPctTurns > 0 && e.DmgReductionPct > 0 {
		dmg -= dmg * e.DmgReductionPct / 100
		if dmg < 1 {
			dmg = 1
		}
	}
	return dmg
}

func (e *TempEffects) ApplyOutgoingDamage(dmg int, consumeHit bool) int {
	if dmg <= 0 {
		return dmg
	}
	if e.OutgoingDmgPct > 0 {
		dmg += dmg * e.OutgoingDmgPct / 100
		if dmg < 1 {
			dmg = 1
		}
	}
	if consumeHit && e.OutgoingDmgHits > 0 {
		e.OutgoingDmgHits--
		if e.OutgoingDmgHits <= 0 {
			e.OutgoingDmgPct = 0
		}
	}
	return dmg
}

func (e *TempEffects) AdvanceTurn() {
	dec := func(turns *int, clear func()) {
		if *turns <= 0 {
			return
		}
		*turns--
		if *turns <= 0 {
			clear()
		}
	}
	dec(&e.CABonusTurns, func() { e.CABonus = 0 })
	dec(&e.CAPenaltyTurns, func() { e.CAPenalty = 0 })
	dec(&e.AtkPenaltyTurns, func() { e.AtkPenalty = 0 })
	dec(&e.DmgReductionPctTurns, func() { e.DmgReductionPct = 0 })
	dec(&e.ForceOneDamageTurns, func() {})
	dec(&e.EnemyDotTurns, func() { e.EnemyDotDmg = 0 })
	dec(&e.SkipEnemyAttackTurns, func() {})
	dec(&e.SkipOwnTurnTurns, func() {})
	dec(&e.MonsterStatusTurns, func() { e.MonsterStatus = "" })
	if e.CritMin > 0 {
		// Crit range buffs valem 1 turno por padrão.
		e.CritMin = 0
	}
}

func (e *TempEffects) SetCABonus(v, turns int) {
	if turns > e.CABonusTurns || v > e.CABonus {
		e.CABonus = v
		e.CABonusTurns = turns
	}
}

func (e *TempEffects) SetCAPenalty(v, turns int) {
	if turns > e.CAPenaltyTurns || v > e.CAPenalty {
		e.CAPenalty = v
		e.CAPenaltyTurns = turns
	}
}

func (e *TempEffects) SetAtkPenalty(v, turns int) {
	if turns > e.AtkPenaltyTurns || v > e.AtkPenalty {
		e.AtkPenalty = v
		e.AtkPenaltyTurns = turns
	}
}

func (e *TempEffects) SetDamageReduction(v, turns int) {
	if turns > e.DmgReductionPctTurns || v > e.DmgReductionPct {
		e.DmgReductionPct = v
		e.DmgReductionPctTurns = turns
	}
}

func (e *TempEffects) SetForceOneDamage(turns int) {
	if turns > e.ForceOneDamageTurns {
		e.ForceOneDamageTurns = turns
	}
}

func (e *TempEffects) SetOutgoingDmg(v, hits int) {
	if hits > e.OutgoingDmgHits || v > e.OutgoingDmgPct {
		e.OutgoingDmgPct = v
		e.OutgoingDmgHits = hits
	}
}

func (e *TempEffects) SetForceCrit(hits int) {
	if hits > e.ForceCritHits {
		e.ForceCritHits = hits
	}
}

func (e *TempEffects) ConsumeForceCrit() bool {
	if e.ForceCritHits <= 0 {
		return false
	}
	e.ForceCritHits--
	return true
}

func (e *TempEffects) SetCritMin(v int) {
	if e.CritMin == 0 || v < e.CritMin {
		e.CritMin = v
	}
}

func (e *TempEffects) SetEnemyDot(dmg, turns int) {
	if turns > e.EnemyDotTurns || dmg > e.EnemyDotDmg {
		e.EnemyDotDmg = dmg
		e.EnemyDotTurns = turns
	}
}

func (e *TempEffects) ApplyEnemyDot() (int, string) {
	if e.EnemyDotTurns <= 0 || e.EnemyDotDmg <= 0 {
		return 0, ""
	}
	dmg := e.EnemyDotDmg
	e.EnemyDotTurns--
	if e.EnemyDotTurns <= 0 {
		e.EnemyDotDmg = 0
	}
	return dmg, fmt.Sprintf("🔥 *Queimadura!* Inimigo sofre *%d* de dano.\n", dmg)
}

func (e *TempEffects) SetSkipEnemyAttack(turns int) {
	if turns > e.SkipEnemyAttackTurns {
		e.SkipEnemyAttackTurns = turns
	}
}

func (e *TempEffects) ConsumeSkipEnemyAttack() bool {
	if e.SkipEnemyAttackTurns <= 0 {
		return false
	}
	e.SkipEnemyAttackTurns--
	return true
}

func (e *TempEffects) SetSkipOwnTurn(turns int) {
	if turns > e.SkipOwnTurnTurns {
		e.SkipOwnTurnTurns = turns
	}
}

func (e *TempEffects) ConsumeSkipOwnTurn() bool {
	if e.SkipOwnTurnTurns <= 0 {
		return false
	}
	e.SkipOwnTurnTurns--
	return true
}

// SetMonsterStatus stores a named status on the monster (for synergy checks).
// Only overwrites if the new duration is longer.
func (e *TempEffects) SetMonsterStatus(status string, turns int) {
	if turns > e.MonsterStatusTurns {
		e.MonsterStatus = status
		e.MonsterStatusTurns = turns
	}
}

// HasMonsterStatus returns true if the monster currently has the given status.
func (e *TempEffects) HasMonsterStatus(status string) bool {
	return e.MonsterStatus == status && e.MonsterStatusTurns > 0
}

var (
	pveEffects = map[int]*TempEffects{} // charID -> efeitos no combate PvE
	pvpEffects = map[int]*TempEffects{} // charID -> efeitos no combate PvP
)

func getPVEEffects(charID int) *TempEffects {
	if fx, ok := pveEffects[charID]; ok {
		return fx
	}
	fx := &TempEffects{}
	pveEffects[charID] = fx
	return fx
}

func resetPVEEffects(charID int) {
	delete(pveEffects, charID)
}

func getPVPEffects(charID int) *TempEffects {
	if fx, ok := pvpEffects[charID]; ok {
		return fx
	}
	fx := &TempEffects{}
	pvpEffects[charID] = fx
	return fx
}

func resetPVPEffects(charID int) {
	delete(pvpEffects, charID)
}

func applySkillEffectsPVE(skillID string, fx *TempEffects) string {
	switch skillID {
	case "w_shield_bash":
		fx.SetCAPenalty(1, 1)
		return "🛡️ CA do inimigo reduzida em 1 por 1 turno."
	case "w_fortress":
		fx.SetCABonus(4, 3)
		fx.SetDamageReduction(30, 3)
		return "🏰 +4 CA e -30% dano recebido por 3 turnos."
	case "w_divine_guard":
		fx.SetForceOneDamage(2)
		return "⛪ Dano recebido reduzido para 1 por 2 turnos."
	case "w_battle_cry":
		fx.SetOutgoingDmg(25, 3)
		return "📣 +25% dano nos próximos 3 golpes."
	case "w_reckless":
		fx.SetOutgoingDmg(40, 1)
		fx.SetCABonus(-2, 1)
		return "💥 +40% dano no próximo golpe e -2 CA por 1 turno."
	case "w_blood_rage":
		fx.SetOutgoingDmg(80, 3)
		return "🩸 +80% dano nos próximos 3 golpes."
	case "w_rampage":
		fx.SetCritMin(18)
		return "🌪️ Crítico ampliado (18-20) neste turno."
	case "w_cleave":
		fx.SetCritMin(19)
		return "🌀 Crítico ampliado (19-20) neste turno."
	case "w_titan_blow":
		fx.SetOutgoingDmg(50, 1)
		return "🏔️ Golpe ignora parte da defesa (+50% dano no próximo golpe)."
	case "m_ice_shard":
		fx.SetCAPenalty(1, 2)
		return "🧊 CA do inimigo reduzida em 1 por 2 turnos."
	case "m_frost_nova":
		fx.SetCAPenalty(3, 2)
		fx.SetAtkPenalty(2, 2)
		return "🌨️ CA do inimigo -3 e ataque inimigo -2 por 2 turnos."
	case "m_arcane_shield":
		fx.SetCABonus(4, 2)
		return "🔮 +4 CA por 2 turnos."
	case "m_flame_wave":
		fx.SetEnemyDot(5, 2)
		return "🔥 Inimigo queimando: 5 dano por 2 turnos."
	case "m_blizzard":
		fx.SetSkipEnemyAttack(1)
		return "🌪️ Inimigo perderá o próximo ataque."
	case "m_absolute_zero":
		fx.SetCAPenalty(5, 1)
		return "🧊 CA do inimigo reduzida em 5 no próximo turno."
	case "m_meteor":
		fx.SetCritMin(18)
		return "☄️ Crítico ampliado (18-20) neste turno."
	case "m_chain_lightning":
		fx.SetOutgoingDmg(20, 1)
		return "🌩️ Precisão arcana: dano aumentado neste ataque."
	case "m_arcane_burst":
		fx.SetOutgoingDmg(100, 1)
		return "💫 Explosão potencializada (+100% dano se escudo ativo)."
	case "r_expose":
		fx.SetCAPenalty(3, 2)
		return "🔍 CA do inimigo reduzida em 3 por 2 turnos."
	case "r_smoke_bomb":
		fx.SetAtkPenalty(4, 2)
		return "💨 Ataque do inimigo reduzido por 2 turnos."
	case "r_shadow_step":
		fx.SetForceCrit(1)
		return "👤 Próximo ataque será crítico garantido."
	case "r_vital_strike":
		fx.SetCritMin(17)
		return "🎯 Crítico ampliado (17-20) neste turno."
	case "a_aimed_shot":
		fx.SetOutgoingDmg(20, 1)
		return "🎯 Tiro focado: dano aumentado neste ataque."
	case "a_headshot":
		fx.SetCritMin(18)
		fx.SetOutgoingDmg(25, 1)
		return "🎖️ Crítico ampliado (18-20) e dano aumentado neste ataque."
	case "a_deadeye":
		fx.SetOutgoingDmg(35, 1)
		return "🌟 Tiro perfeito: dano aumentado neste ataque."
	case "a_frost_arrow":
		fx.SetCAPenalty(2, 2)
		return "🧊 CA do inimigo reduzida em 2 por 2 turnos."
	case "a_multishot":
		fx.SetOutgoingDmg(120, 1)
		return "🌧️ Rajada tripla simulada: dano amplificado."
	case "a_volley":
		fx.SetOutgoingDmg(220, 1)
		return "⛈️ Saraivada simulada: dano amplificado."
	case "a_quick_shot":
		return "🏹 Tiro rápido pronto."
	}
	return ""
}

func applySkillEffectsPVP(skillID string, self, target *TempEffects) string {
	switch skillID {
	case "w_shield_bash":
		target.SetCAPenalty(1, 1)
		return "🛡️ CA do alvo reduzida em 1 por 1 turno."
	case "w_fortress":
		self.SetCABonus(4, 3)
		self.SetDamageReduction(30, 3)
		return "🏰 +4 CA e -30% dano recebido por 3 turnos."
	case "w_divine_guard":
		self.SetForceOneDamage(2)
		return "⛪ Dano recebido reduzido para 1 por 2 turnos."
	case "w_battle_cry":
		self.SetOutgoingDmg(25, 3)
		return "📣 +25% dano nos próximos 3 golpes."
	case "w_reckless":
		self.SetOutgoingDmg(40, 1)
		self.SetCABonus(-2, 1)
		return "💥 +40% dano no próximo golpe e -2 CA por 1 turno."
	case "w_blood_rage":
		self.SetOutgoingDmg(80, 3)
		return "🩸 +80% dano nos próximos 3 golpes."
	case "w_rampage":
		self.SetCritMin(18)
		return "🌪️ Crítico ampliado (18-20) neste turno."
	case "w_cleave":
		self.SetCritMin(19)
		return "🌀 Crítico ampliado (19-20) neste turno."
	case "w_titan_blow":
		self.SetOutgoingDmg(50, 1)
		return "🏔️ Golpe ignora parte da defesa (+50% dano no próximo golpe)."
	case "m_ice_shard":
		target.SetCAPenalty(1, 2)
		return "🧊 CA do alvo reduzida em 1 por 2 turnos."
	case "m_frost_nova":
		target.SetCAPenalty(3, 2)
		target.SetAtkPenalty(2, 2)
		return "🌨️ CA do alvo -3 e ataque -2 por 2 turnos."
	case "m_arcane_shield":
		self.SetCABonus(4, 2)
		return "🔮 +4 CA por 2 turnos."
	case "m_flame_wave":
		target.SetEnemyDot(5, 2)
		return "🔥 Alvo queimando: 5 dano por 2 turnos."
	case "m_blizzard":
		target.SetSkipOwnTurn(1)
		return "🌪️ Alvo perderá o próximo turno."
	case "m_absolute_zero":
		target.SetCAPenalty(5, 1)
		return "🧊 CA do alvo reduzida em 5 no próximo turno."
	case "m_meteor":
		self.SetCritMin(18)
		return "☄️ Crítico ampliado (18-20) neste turno."
	case "m_chain_lightning":
		self.SetOutgoingDmg(20, 1)
		return "🌩️ Precisão arcana: dano aumentado neste ataque."
	case "m_arcane_burst":
		self.SetOutgoingDmg(100, 1)
		return "💫 Explosão potencializada (+100% dano se escudo ativo)."
	case "r_expose":
		target.SetCAPenalty(3, 2)
		return "🔍 CA do alvo reduzida em 3 por 2 turnos."
	case "r_smoke_bomb":
		target.SetAtkPenalty(4, 2)
		return "💨 Ataque do alvo reduzido por 2 turnos."
	case "r_shadow_step":
		self.SetForceCrit(1)
		return "👤 Próximo ataque será crítico garantido."
	case "r_vital_strike":
		self.SetCritMin(17)
		return "🎯 Crítico ampliado (17-20) neste turno."
	case "a_aimed_shot":
		self.SetOutgoingDmg(20, 1)
		return "🎯 Tiro focado: dano aumentado neste ataque."
	case "a_headshot":
		self.SetCritMin(18)
		self.SetOutgoingDmg(25, 1)
		return "🎖️ Crítico ampliado (18-20) e dano aumentado neste ataque."
	case "a_deadeye":
		self.SetOutgoingDmg(35, 1)
		return "🌟 Tiro perfeito: dano aumentado neste ataque."
	case "a_frost_arrow":
		target.SetCAPenalty(2, 2)
		return "🧊 CA do alvo reduzida em 2 por 2 turnos."
	case "a_multishot":
		self.SetOutgoingDmg(120, 1)
		return "🌧️ Rajada tripla simulada: dano amplificado."
	case "a_volley":
		self.SetOutgoingDmg(220, 1)
		return "⛈️ Saraivada simulada: dano amplificado."
	case "a_quick_shot":
		return "🏹 Tiro rápido pronto."
	}
	return ""
}

func formatEffectMsg(msg string) string {
	if msg == "" {
		return ""
	}
	return fmt.Sprintf("%s\n", msg)
}
