import { readdirSync, readFileSync, existsSync } from 'node:fs'
import { join, dirname } from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'
import {
  getSettlementDisplay,
  validateSettlementState,
} from '../src/composables/yuzhousha/settlementDisplay'
import type { YuzhoushaState } from '../src/types/yuzhousha'

const here = dirname(fileURLToPath(import.meta.url))
const FIXTURE_DIR = join(here, 'fixtures', 'yzs', 'settlements')

interface SettlementFixtureFile {
  meta?: {
    mode?: string
    seed?: number
    label?: string
    winner_team?: number
  }
  state: YuzhoushaState
}

function loadFixtures(): { name: string; file: SettlementFixtureFile }[] {
  if (!existsSync(FIXTURE_DIR)) {
    return []
  }
  return readdirSync(FIXTURE_DIR)
    .filter((f) => f.endsWith('.json'))
    .map((name) => {
      const raw = readFileSync(join(FIXTURE_DIR, name), 'utf8')
      return { name, file: JSON.parse(raw) as SettlementFixtureFile }
    })
}

describe('settlementDisplay unit', () => {
  it('finished state shows message as centerHint', () => {
    const state: YuzhoushaState = {
      id: 't',
      phase: 'finished',
      turn_step: 'play',
      current_turn: 0,
      human_player: 0,
      players: [
        { index: 0, name: '主公', is_ai: true, hp: 0, max_hp: 4, hand_count: 0 },
        { index: 1, name: '反贼1', is_ai: true, hp: 2, max_hp: 4, hand_count: 1 },
        { index: 2, name: '内奸', is_ai: true, hp: 2, max_hp: 3, hand_count: 0 },
        { index: 3, name: '反贼2', is_ai: true, hp: 3, max_hp: 4, hand_count: 2 },
      ],
      message: '主公阵亡，反贼获胜',
      winner_index: 3,
      winner_team: 1,
      mode: 'identity_5',
      draw_count: 0,
      discard_count: 10,
      turn_deadline_unix: 0,
      events: [{ type: 'game_over', message: '主公阵亡，反贼获胜' }],
    }
    const d = getSettlementDisplay(state)
    expect(d.isFinished).toBe(true)
    expect(d.centerHint).toBe('主公阵亡，反贼获胜')
    expect(d.showRestart).toBe(true)
    expect(validateSettlementState(state)).toEqual([])
  })
})

describe('settlement fixtures (from backend sim harvest)', () => {
  const fixtures = loadFixtures()

  it('fixture directory exists or skip with hint', () => {
    if (fixtures.length === 0) {
      console.warn(
        'No fixtures in test/fixtures/yzs/settlements — run: cd backend && CARD_SIM=1 CARD_UI_FIXTURE=1 go test -tags cardtest ./test/yuzhousha/... -run TestHarvestYzsSettlementFixtures -v',
      )
    }
    expect(true).toBe(true)
  })

  for (const { name, file } of fixtures) {
    it(`validates ${name}`, () => {
      const errors = validateSettlementState(file.state)
      if (errors.length) {
        console.error(name, errors, file.meta)
      }
      expect(errors, name).toEqual([])
      if (file.meta?.winner_team != null && file.state.winner_team != null) {
        expect(file.state.winner_team).toBe(file.meta.winner_team)
      }
      if (file.meta?.mode && file.state.mode) {
        expect(file.state.mode).toBe(file.meta.mode)
      }
    })
  }
})
