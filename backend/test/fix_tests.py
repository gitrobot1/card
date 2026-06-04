#!/usr/bin/env python3
"""Fix double prefixes and broken patterns after convert_tests.py."""

import re
from pathlib import Path

ROOT = Path(__file__).parent


def fix_content(text: str) -> str:
    subs = [
        (r"uno\.uno\.", "uno."),
        (r"engine\.engine\.", "engine."),
        (r"douniu\.douniu\.", "douniu."),
        (r"doudizhu\.doudizhu\.", "doudizhu."),
        (r"zhajinhua\.zhajinhua\.", "zhajinhua."),
        (r"skill\.engine\.engine\.", "skill."),
        (r"skill\.engine\.", "skill."),
        (r"\buno\.uno\.Card:", "Card:"),
        (r"\.uno\.uno\.Card\b", ".Card"),
        (r"g\.SetRollRoundSum\((\d+)\]\s*=", r"g.SetRollRoundSum(\1, "),
        (r"g\.SetRollRoundSum\((\d+)\]\s*!=", r"g.RollRoundSum(\1) !="),
        (r"g\.engine\.", "g."),
        (r"g2\.runSkillHooks\(", "g2.RunSkillHooks("),
        (r"g2\.syncCounts\(\)", "g2.SyncCounts()"),
        (r"g2\.canBingliangTarget\(", "g2.CanBingliangTargetForTest("),
        (r"g\.Players\[(\d+)\]\.engine\.engine\.", r"g.Players[\1]."),
    ]
    for pat, repl in subs:
        text = re.sub(pat, repl, text)
    return text


def qualified_replace(text: str, name: str, prefix: str) -> str:
    """Replace bare identifier, skip if already package-qualified."""
    pat = rf"(?<!{re.escape(prefix)}\.)\b{re.escape(name)}\b"
    return re.sub(pat, f"{prefix}.{name}", text)


def main() -> None:
    for f in ROOT.rglob("*_test.go"):
        text = fix_content(f.read_text())
        f.write_text(text)
        print("fixed", f)


if __name__ == "__main__":
    main()
