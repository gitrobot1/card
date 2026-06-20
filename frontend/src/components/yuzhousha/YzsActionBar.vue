<script setup lang="ts">
import { useYzsGameInject } from '../../composables/yuzhousha/useYzsGame'

const fanjianSuits = [
  { code: 'H', label: '红桃' },
  { code: 'D', label: '方块' },
  { code: 'S', label: '黑桃' },
  { code: 'C', label: '梅花' },
] as const

const {
  activatableSkillIds,
  canCancelWusheng,
  canSubmitBagua,
  canSubmitCancel,
  canSubmitEndTurn,
  canSubmitFankui,
  canSubmitGanglieDiscard,
  canSubmitDdzJudgeCancel,
  canSubmitGuicai,
  canSubmitGuidao,
  canSubmitLeiji,
  canSubmitLiuli,
  canSubmitPlay,
  canSubmitQixi,
  canSubmitTianxiang,
  canSubmitTuxi,
  canSubmitYijiGive,
  canSubmitYinghunDiscard,
  discardNeeded,
  isAnimating,
  isDyingRescue,
  isFanjianSuit,
  isFankui,
  isGanglieChoice,
  isDdzJudgeCancel,
  isGanglieOffer,
  isGuanYuFollow,
  isGuicai,
  isGuidao,
  isJianxiong,
  isJijiHeal,
  isJijiangRespond,
  isLeijiOffer,
  isLiuliOffer,
  isLuanwu,
  isMyDiscard,
  isMyDraw,
  isMyPlay,
  isMyPrepare,
  isMyResponse,
  isPeekDeck,
  isQilinBow,
  isQixiTake,
  isGuoHeTake,
  isTanNangTake,
  isTakeWindow,
  selectedTargetZone,
  selectedTargetCardId,
  isSkillOnlyResponse,
  isTianxiangOffer,
  isTuxiTake,
  isWuguPick,
  isWuxiekOffer,
  isYijiGive,
  isYijiOffer,
  isYinghunChoice,
  isYinghunDiscard,
  loading,
  selectedDiscardIds,
  showActionButton,
  submitBaguaJudge,
  submitCancelResponse,
  submitCancelWusheng,
  submitEndTurn,
  submitFanjianSuit,
  submitGanglieDiscard,
  submitGanglieTakeDamage,
  submitPassDraw,
  submitPassPrepare,
  submitPassYijiGive,
  submitPlayCard,
  submitSkill,
  submitTuxiSkip,
  submitYinghunDiscard,
  submitYinghunOption
} = useYzsGameInject()
</script>


<template>
  <div v-if="showActionButton" class="yzs__action-float">
    <template v-if="isMyDiscard">
      <button
        type="button"
        class="ddz__btn"
        :class="{ 'ddz__btn--primary': canSubmitPlay }"
        :disabled="!canSubmitPlay"
        @click="submitPlayCard"
      >
        弃牌<span v-if="discardNeeded > 0" class="yzs__discard-count">（{{ selectedDiscardIds.length }}/{{ discardNeeded }}）</span>
      </button>
    </template>

    <template v-else-if="isPeekDeck">
      <button
        type="button"
        class="ddz__btn ddz__btn--primary"
        :disabled="!canSubmitPlay"
        @click="submitPlayCard"
      >
        确认观星
      </button>
    </template>

    <template v-else-if="isMyPrepare">
      <button
        v-if="activatableSkillIds.has('guanxing')"
        type="button"
        class="ddz__btn"
        :disabled="loading || isAnimating"
        @click="submitSkill('guanxing')"
      >
        观星
      </button>
      <button
        v-if="activatableSkillIds.has('luoshen')"
        type="button"
        class="ddz__btn"
        :disabled="loading || isAnimating"
        @click="submitSkill('luoshen')"
      >
        洛神
      </button>
      <button
        v-if="activatableSkillIds.has('yinghun')"
        type="button"
        class="ddz__btn"
        :disabled="loading || isAnimating"
        @click="submitSkill('yinghun')"
      >
        英魂
      </button>
      <button
        v-if="activatableSkillIds.has('hunzi')"
        type="button"
        class="ddz__btn ddz__btn--primary"
        :disabled="loading || isAnimating"
        @click="submitSkill('hunzi')"
      >
        魂姿
      </button>
      <button
        type="button"
        class="ddz__btn"
        :disabled="loading || isAnimating"
        @click="submitPassPrepare"
      >
        跳过
      </button>
    </template>

    <template v-else-if="isMyDraw">
      <button
        v-if="activatableSkillIds.has('luoyi')"
        type="button"
        class="ddz__btn"
        :disabled="loading || isAnimating"
        @click="submitSkill('luoyi')"
      >
        裸衣
      </button>
      <template v-if="activatableSkillIds.has('tuxi')">
        <button
          type="button"
          class="ddz__btn"
          :disabled="loading || isAnimating"
          @click="submitTuxiSkip(1)"
        >
          突袭·少摸1
        </button>
        <button
          type="button"
          class="ddz__btn"
          :disabled="loading || isAnimating"
          @click="submitTuxiSkip(2)"
        >
          突袭·少摸2
        </button>
      </template>
      <button
        v-if="activatableSkillIds.has('shuangxiong')"
        type="button"
        class="ddz__btn"
        :disabled="loading || isAnimating"
        @click="submitSkill('shuangxiong')"
      >
        双雄
      </button>
      <button
        type="button"
        class="ddz__btn ddz__btn--primary"
        :disabled="loading || isAnimating"
        @click="submitPassDraw"
      >
        摸牌
      </button>
    </template>

    <template v-else-if="isMyResponse">
      <button
        v-if="canCancelWusheng"
        type="button"
        class="ddz__btn"
        @click="submitCancelWusheng"
      >
        取消武圣
      </button>

      <template v-if="isGanglieChoice">
        <button
          type="button"
          class="ddz__btn ddz__btn--primary"
          :disabled="!canSubmitGanglieDiscard"
          @click="submitGanglieDiscard"
        >
          弃2张
        </button>
        <button
          type="button"
          class="ddz__btn"
          :disabled="loading || isAnimating"
          @click="submitGanglieTakeDamage"
        >
          受1点伤害
        </button>
      </template>

      <template v-else-if="isDdzJudgeCancel">
        <button
          type="button"
          class="ddz__btn ddz__btn--primary"
          :disabled="!canSubmitDdzJudgeCancel"
          @click="submitSkill('ddz_judge_cancel')"
        >
          弃2张取消判定
        </button>
        <button
          v-if="canSubmitCancel"
          type="button"
          class="ddz__btn"
          @click="submitCancelResponse"
        >
          取消
        </button>
      </template>

      <template v-else-if="isYijiGive">
        <button
          type="button"
          class="ddz__btn ddz__btn--primary"
          :disabled="!canSubmitYijiGive"
          @click="submitSkill('yiji')"
        >
          给出
        </button>
        <button
          type="button"
          class="ddz__btn"
          :disabled="loading || isAnimating"
          @click="submitPassYijiGive"
        >
          完成
        </button>
      </template>

      <template v-else-if="isFanjianSuit">
        <div class="yzs__action-float-btns">
          <button
            v-for="suit in fanjianSuits"
            :key="suit.code"
            type="button"
            class="ddz__btn"
            :disabled="loading || isAnimating"
            @click="submitFanjianSuit(suit.code)"
          >
            {{ suit.label }}
          </button>
        </div>
      </template>

      <template v-else-if="isYinghunChoice">
        <button
          type="button"
          class="ddz__btn"
          :disabled="loading || isAnimating"
          @click="submitYinghunOption('opp_draw_x_discard_1')"
        >
          令对手摸 X 弃 1
        </button>
        <button
          type="button"
          class="ddz__btn"
          :disabled="loading || isAnimating"
          @click="submitYinghunOption('opp_draw_1_discard_x')"
        >
          令对手摸 1 弃 X
        </button>
      </template>

      <template v-else-if="isYinghunDiscard">
        <button
          type="button"
          class="ddz__btn ddz__btn--primary"
          :disabled="!canSubmitYinghunDiscard"
          @click="submitYinghunDiscard"
        >
          弃牌
        </button>
        <button
          v-if="canSubmitCancel"
          type="button"
          class="ddz__btn"
          @click="submitCancelResponse"
        >
          取消
        </button>
      </template>

      <template v-else-if="isGanglieOffer">
        <button
          type="button"
          class="ddz__btn ddz__btn--primary"
          :disabled="loading || isAnimating"
          @click="submitSkill('ganglie')"
        >
          刚烈
        </button>
        <button
          v-if="canSubmitCancel"
          type="button"
          class="ddz__btn"
          @click="submitCancelResponse"
        >
          取消
        </button>
      </template>

      <template v-else-if="isYijiOffer">
        <button
          type="button"
          class="ddz__btn ddz__btn--primary"
          :disabled="loading || isAnimating"
          @click="submitSkill('yiji')"
        >
          遗计
        </button>
        <button
          v-if="canSubmitCancel"
          type="button"
          class="ddz__btn"
          @click="submitCancelResponse"
        >
          取消
        </button>
      </template>

      <template v-else-if="isJianxiong">
        <button
          type="button"
          class="ddz__btn ddz__btn--primary"
          :disabled="loading || isAnimating"
          @click="submitSkill('jianxiong')"
        >
          奸雄
        </button>
        <button
          v-if="canSubmitCancel"
          type="button"
          class="ddz__btn"
          @click="submitCancelResponse"
        >
          取消
        </button>
      </template>

      <template v-else-if="isLeijiOffer">
        <button
          type="button"
          class="ddz__btn ddz__btn--primary"
          :disabled="!canSubmitLeiji"
          @click="submitSkill('leiji')"
        >
          雷击
        </button>
        <button
          v-if="canSubmitCancel"
          type="button"
          class="ddz__btn"
          @click="submitCancelResponse"
        >
          取消
        </button>
      </template>

      <template v-else-if="isFankui">
        <button
          type="button"
          class="ddz__btn ddz__btn--primary"
          :disabled="!canSubmitFankui"
          @click="submitSkill('fankui')"
        >
          反馈
        </button>
        <button
          v-if="canSubmitCancel"
          type="button"
          class="ddz__btn"
          @click="submitCancelResponse"
        >
          取消
        </button>
      </template>

      <template v-else-if="isTuxiTake">
        <button
          type="button"
          class="ddz__btn ddz__btn--primary"
          :disabled="!canSubmitTuxi"
          @click="submitSkill('tuxi')"
        >
          突袭
        </button>
        <button
          v-if="canSubmitCancel"
          type="button"
          class="ddz__btn"
          @click="submitCancelResponse"
        >
          取消
        </button>
      </template>

      <template v-else-if="isQixiTake">
        <button
          type="button"
          class="ddz__btn ddz__btn--primary"
          :disabled="!canSubmitQixi"
          @click="submitSkill('qixi')"
        >
          奇袭
        </button>
        <button
          v-if="canSubmitCancel"
          type="button"
          class="ddz__btn"
          @click="submitCancelResponse"
        >
          取消
        </button>
      </template>

      <template v-else-if="isTakeWindow">
        <button
          type="button"
          class="ddz__btn ddz__btn--primary"
          :disabled="selectedTargetZone === '' || selectedTargetCardId === ''"
          @click="submitSkill('')"
        >
          {{ isGuoHeTake ? '拆掉' : '获得' }}
        </button>
        <button
          v-if="canSubmitCancel"
          type="button"
          class="ddz__btn"
          @click="submitCancelResponse"
        >
          取消
        </button>
      </template>

      <template v-else-if="isGuicai">
        <button
          type="button"
          class="ddz__btn ddz__btn--primary"
          :disabled="!canSubmitGuicai"
          @click="submitSkill('guicai')"
        >
          鬼才
        </button>
        <button
          v-if="canSubmitCancel"
          type="button"
          class="ddz__btn"
          @click="submitCancelResponse"
        >
          取消
        </button>
      </template>

      <template v-else-if="isGuidao">
        <button
          type="button"
          class="ddz__btn ddz__btn--primary"
          :disabled="!canSubmitGuidao"
          @click="submitSkill('guidao')"
        >
          鬼道
        </button>
        <button
          v-if="canSubmitCancel"
          type="button"
          class="ddz__btn"
          @click="submitCancelResponse"
        >
          取消
        </button>
      </template>

      <template v-else-if="isTianxiangOffer">
        <button
          type="button"
          class="ddz__btn ddz__btn--primary"
          :disabled="!canSubmitTianxiang"
          @click="submitSkill('tianxiang')"
        >
          天香
        </button>
        <button
          v-if="canSubmitCancel"
          type="button"
          class="ddz__btn"
          @click="submitCancelResponse"
        >
          取消
        </button>
      </template>

      <template v-else-if="isLiuliOffer">
        <button
          type="button"
          class="ddz__btn ddz__btn--primary"
          :disabled="!canSubmitLiuli"
          @click="submitSkill('liuli')"
        >
          流离
        </button>
        <button
          v-if="canSubmitCancel"
          type="button"
          class="ddz__btn"
          @click="submitCancelResponse"
        >
          取消
        </button>
      </template>

      <template v-else>
        <button
          v-if="isQilinBow"
          type="button"
          class="ddz__btn ddz__btn--primary"
          :disabled="!canSubmitPlay"
          @click="submitPlayCard"
        >
          弃坐骑
        </button>
        <button
          v-else-if="isWuguPick || isGuanYuFollow || isJijiangRespond || isLuanwu || isDyingRescue || isWuxiekOffer || !isSkillOnlyResponse"
          type="button"
          class="ddz__btn"
          :class="{ 'ddz__btn--primary': canSubmitPlay }"
          :disabled="!canSubmitPlay"
          @click="submitPlayCard"
        >
          {{ isWuguPick ? '选牌' : '出牌' }}
        </button>
        <button
          v-if="canSubmitBagua"
          type="button"
          class="ddz__btn"
          @click="submitBaguaJudge"
        >
          八卦判定
        </button>
        <!-- 在所有响应阶段，统一显示“取消”按钮（无懈可击等无法出牌时也能跳过） -->
        <button
          v-if="!loading && !isAnimating && !isFanjianSuit && !isYinghunChoice && !isYinghunDiscard && !isGanglieChoice"
          type="button"
          class="ddz__btn"
          @click="submitCancelResponse"
        >
          取消
        </button>
      </template>
    </template>

    <template v-else-if="isMyPlay">
      <button
        v-if="activatableSkillIds.has('luanwu')"
        type="button"
        class="ddz__btn"
        :disabled="loading || isAnimating"
        @click="submitSkill('luanwu')"
      >
        乱武
      </button>
      <button
        type="button"
        class="ddz__btn"
        :class="{ 'ddz__btn--primary': canSubmitPlay }"
        :disabled="!canSubmitPlay"
        @click="submitPlayCard"
      >
        出牌
      </button>
      <button
        type="button"
        class="ddz__btn"
        :disabled="!canSubmitEndTurn"
        @click="submitEndTurn"
      >
        结束出牌
      </button>
    </template>

    <template v-else-if="isJijiHeal">
      <button
        type="button"
        class="ddz__btn ddz__btn--primary"
        :disabled="!canSubmitPlay"
        @click="submitPlayCard"
      >
        出牌
      </button>
    </template>
  </div>
</template>
