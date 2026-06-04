#!/usr/bin/env python3
"""Convert in-package _test.go sources to external test packages under test/."""

import re
from pathlib import Path

ROOT = Path(__file__).parent


def qreplace(text: str, name: str, prefix: str) -> str:
    pat = rf"(?<!{re.escape(prefix)}\.)(?<!\w)\b{re.escape(name)}\b"
    return re.sub(pat, f"{prefix}.{name}", text)


def convert_yuzhousha(text: str) -> str:
    text = re.sub(r"^package engine\s*\n", "", text)
    text = re.sub(r"^import \([\s\S]*?\)\s*\n", "", text)
    text = re.sub(r"^import \"[^\"]+\"\s*\n", "", text)
    text = text.lstrip("\n")

    pkg = "engine"
    for name in [
        "NewSolo1v1", "HeroesCatalog", "RunAIActions", "NewBasicDeck",
        "PlayCard", "PlayCardWithTarget", "PassResponse", "RespondCard",
        "RespondShan", "RespondWuxiek", "EndPlay", "DiscardCards", "UseSkill",
        "TryBaguaJudge", "ApplyTieqi", "ApplyGuicaiReplace", "FankuiTakeFrom",
        "StartLuoshen", "StartPeekDeck", "FinishPeekDeck", "PassPrepare",
        "PickWuguCard", "QilinDiscardHorse",
        "PhasePlaying", "PhaseResponse", "PhaseFinished",
        "StepPlay", "StepDraw", "StepDiscard", "StepPrepare",
        "CardSha", "CardShan", "CardTao", "CardJiu", "CardGuoHe", "CardTanNang",
        "CardNanMan", "CardWanJian", "CardJueDou", "CardLeBu", "CardBingLiang",
        "CardShanDian", "CardWuGu", "CardTaoYuan", "CardWuZhong", "CardWuxiek",
        "CardWeapon1", "CardWeapon2", "CardWeapon3", "CardWeapon4", "CardWeapon5",
        "CardArmor", "CardPlusHorse", "CardMinusHorse",
        "EquipWeapon", "EquipArmor", "EquipPlusHorse", "EquipMinusHorse",
        "ResponseModeSkillFankui", "ResponseModeSkillGuicai", "ResponseModePeekDeck",
        "ResponseModeWuxiekTrick", "ResponseModeWuxiekLebu", "ResponseModeWuguPick",
        "InitialHandSize", "DrawPerTurn", "DefaultMaxHP",
        "CharLiuBei", "CharGuanYu", "CharZhangFei", "CharZhaoYun", "CharZhugeLiang",
        "CharMaChao", "CharHuangYueying", "CharSimaYi", "CharZhenJi",
        "SkillWusheng", "SkillTieqi", "SkillRende",
        "ErrInvalidTarget", "ErrInvalidCard", "ErrInvalidDiscardCount",
        "ErrWrongPhase", "ErrNotYourTurn",
    ]:
        text = qreplace(text, name, pkg)

    # types: only when used as type name (heuristic: after [] or * or space before var)
    for typ in ["Card", "GameEvent", "PlayTarget", "PendingCombat", "UseSkillRequest",
                "PeekDeckRequest", "Game", "Player", "Character", "SkillMeta", "SkillKindPassive"]:
        text = qreplace(text, typ, pkg)

    text = re.sub(r"\bg\.syncCounts\(\)", "g.SyncCounts()", text)
    text = re.sub(r"\bg\.canUseSha\(", "g.CanUseSha(", text)
    text = re.sub(r"\bg\.cardPlaysAs\(", "g.CardPlaysAsForTest(", text)
    text = re.sub(r"\bg\.targetBlockedBySkill\(", "g.TargetBlockedBySkillForTest(", text)
    text = re.sub(r"\bg\.playSha\(", "g.PlaySha(", text)
    text = re.sub(r"\bg\.runSkillHooks\(", "g.RunSkillHooks(", text)
    text = re.sub(r"\bg\.applyDamage\(", "g.ApplyDamageForTest(", text)
    text = re.sub(r"\bg\.notifyInstantTrickUsed\(", "g.NotifyInstantTrickUsedForTest(", text)
    text = re.sub(r"\bg\.beginTurn\(", "g.BeginTurnForTest(", text)
    text = re.sub(r"\bg\.canBingliangTarget\(", "g.CanBingliangTargetForTest(", text)
    text = re.sub(r"\bg\.distanceBetween\(", "g.DistanceBetween(", text)
    text = re.sub(r"\bg\.engine\.", "g.", text)

    extra = '\t"errors"\n' if "errors.Is" in text else ""
    header = (
        "package engine_test\n\nimport (\n"
        f"{extra}\t\"testing\"\n\n\t"
        'engine "github.com/time/card/backend/internal/game/yuzhousha/engine"\n\t'
        '"github.com/time/card/backend/internal/game/yuzhousha/skill"\n)\n\n'
    )
    return header + text


def convert_uno(text: str) -> str:
    text = re.sub(r"^package uno\s*\n", "", text)
    text = re.sub(r"^import \([\s\S]*?\)\s*\n", "", text)
    text = re.sub(r"^import \"testing\"\s*\n", "", text)
    text = text.lstrip("\n")
    pkg = "uno"
    for name in [
        "NewSoloGame", "NewMultiGame", "RunAIActions", "FilterEventsForSeat",
        "PhaseRollForFirst", "PhasePlaying", "PhaseFinished", "InitialHand",
        "RollRound", "EventRollDice", "EventRollTie", "EventFirstPlayer",
    ]:
        text = qreplace(text, name, pkg)
    for typ in ["GameEvent", "Card", "Game"]:
        text = qreplace(text, typ, pkg)
    text = re.sub(r"\bg\.canPlayCard\(", "g.CanPlayCardForTest(", text)
    text = re.sub(r"\bg\.syncCounts\(\)", "g.SyncCountsForTest()", text)
    text = re.sub(r"\bg\.finalizeRollRound\(", "g.FinalizeRollRoundForTest(", text)
    text = re.sub(r"\bg\.rollRoundSums\[(\d+)\]\s*=", r"g.SetRollRoundSum(\1, ", text)
    text = re.sub(r"\bg\.rollRoundSums\[(\d+)\]", r"g.RollRoundSum(\1)", text)
    text = re.sub(r"\bg\.rollContenders\b", "g.RollContenders()", text)
    text = re.sub(r"\bg\.checkAfterElimination\(", "g.CheckAfterEliminationForTest(", text)
    header = (
        "package uno_test\n\nimport (\n\t\"testing\"\n\n\t"
        f'uno "github.com/time/card/backend/internal/game/uno"\n)\n\n'
    )
    return header + text


def convert_douniu(text: str) -> str:
    text = re.sub(r"^package douniu\s*\n", "", text)
    text = re.sub(r"^import \([\s\S]*?\)\s*\n", "", text)
    text = text.lstrip("\n")
    for name in ["AnalyzeHand", "HandNiu9", "HandFiveFlower", "HandFiveSmall", "HandBomb", "HandNiuNiu"]:
        text = qreplace(text, name, "douniu")
    header = (
        "package douniu_test\n\nimport (\n\t\"testing\"\n\n\t"
        f'douniu "github.com/time/card/backend/internal/game/douniu"\n\t'
        f'"github.com/time/card/backend/internal/game/card"\n)\n\n'
    )
    return header + text


def convert_doudizhu(text: str) -> str:
    text = re.sub(r"^package doudizhu\s*\n", "", text)
    text = re.sub(r"^import \([\s\S]*?\)\s*\n", "", text)
    text = text.lstrip("\n")
    text = re.sub(r"\bdebugDealOneCard\b", "doudizhu.DebugDealOneCardEnabled()", text)
    for name in ["NewGame", "PhasePlaying", "PhaseFinished"]:
        text = qreplace(text, name, "doudizhu")
    header = (
        "package doudizhu_test\n\nimport (\n\t\"testing\"\n\n\t"
        f'doudizhu "github.com/time/card/backend/internal/game/doudizhu"\n\t'
        f'"github.com/time/card/backend/internal/game/card"\n)\n\n'
    )
    return header + text


def convert_zhajinhua(text: str) -> str:
    text = re.sub(r"^package zhajinhua\s*\n", "", text)
    text = re.sub(r"^import \([\s\S]*?\)\s*\n", "", text)
    text = text.lstrip("\n")
    for name in [
        "AnalyzeHand", "HandLeopard", "HandStraightFlush", "HandFlush",
        "HandStraight", "HandPair", "HandHigh", "Compare",
    ]:
        text = qreplace(text, name, "zhajinhua")
    header = (
        "package zhajinhua_test\n\nimport (\n\t\"testing\"\n\n\t"
        f'zhajinhua "github.com/time/card/backend/internal/game/zhajinhua"\n\t'
        f'"github.com/time/card/backend/internal/game/card"\n)\n\n'
    )
    return header + text


CONVERTERS = {
    "yuzhousha": convert_yuzhousha,
    "uno": convert_uno,
    "douniu": convert_douniu,
    "doudizhu": convert_doudizhu,
    "zhajinhua": convert_zhajinhua,
}


def main() -> None:
    for game, fn in CONVERTERS.items():
        for f in (ROOT / game).glob("*_test.go"):
            if f.name == "helpers_test.go":
                continue
            raw = f.read_text()
            if raw.startswith("package ") and "_test" in raw.split("\n", 1)[0]:
                continue  # already converted
            f.write_text(fn(raw))
            print("converted", f)
    # helpers always external
    helpers = ROOT / "uno" / "helpers_test.go"
    if helpers.exists():
        helpers.write_text(convert_uno(helpers.read_text()) if "package uno\n" in helpers.read_text() else helpers.read_text())


if __name__ == "__main__":
    main()
