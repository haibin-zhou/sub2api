<template>
  <AppLayout>
    <div class="space-y-6">
      <!-- Loading State -->
      <div v-if="loading" class="flex justify-center py-12">
        <div
          class="h-8 w-8 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"
        ></div>
      </div>

      <!-- Empty State -->
      <div v-else-if="subscriptions.length === 0" class="card p-12 text-center">
        <div
          class="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-gray-100 dark:bg-dark-700"
        >
          <Icon name="creditCard" size="xl" class="text-gray-400" />
        </div>
        <h3 class="mb-2 text-lg font-semibold text-gray-900 dark:text-white">
          {{ t('userSubscriptions.noActiveSubscriptions') }}
        </h3>
        <p class="text-gray-500 dark:text-dark-400">
          {{ t('userSubscriptions.noActiveSubscriptionsDesc') }}
        </p>
      </div>

      <!-- Subscriptions Grid -->
      <div v-else class="grid gap-6 lg:grid-cols-2">
        <div
          v-for="subscription in subscriptions"
          :key="subscription.id"
          class="overflow-hidden rounded-2xl border bg-white dark:bg-dark-800"
          :class="platformBorderClass(subscription.group?.platform || '')"
        >
          <!-- Header -->
          <div
            class="flex items-center justify-between border-b border-gray-100 p-4 dark:border-dark-700"
          >
            <div class="flex items-center gap-3">
              <div :class="['h-1.5 w-1.5 shrink-0 rounded-full', platformAccentDotClass(subscription.group?.platform || '')]" />
              <div>
                <div class="flex items-center gap-2">
                  <h3 class="font-semibold text-gray-900 dark:text-white">
                    {{ subscription.group?.name || `Group #${subscription.group_id}` }}
                  </h3>
                  <span :class="['rounded-md border px-2 py-0.5 text-[11px] font-medium', platformBadgeClass(subscription.group?.platform || '')]">
                    {{ platformLabel(subscription.group?.platform || '') }}
                  </span>
                </div>
                <p v-if="subscription.group?.description" class="mt-0.5 text-xs text-gray-500 dark:text-dark-400">
                  {{ subscription.group.description }}
                </p>
                <div class="mt-1 flex flex-wrap gap-x-3 gap-y-1 text-[11px] text-gray-400 dark:text-gray-500">
                  <span>{{ t('payment.planCard.rate') }}: ×{{ subscription.group?.rate_multiplier ?? 1 }}</span>
                  <span v-if="subscriptionHasPeakRate(subscription)" class="text-amber-700 dark:text-amber-300">
                    {{ t('payment.planCard.peakRate') }}: {{ subscriptionPeakRateLabel(subscription) }}
                  </span>
                </div>
              </div>
            </div>
            <div class="flex items-center gap-2">
              <span
                :class="[
                  'rounded-full px-2 py-0.5 text-xs font-medium',
                  subscription.status === 'active'
                    ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-300'
                    : subscription.status === 'expired'
                      ? 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-gray-400'
                      : 'bg-red-100 text-red-700 dark:bg-red-900/40 dark:text-red-300'
                ]"
              >
                {{ t(`userSubscriptions.status.${subscription.status}`) }}
              </span>
              <button
                v-if="subscription.status === 'active'"
                :class="['rounded-lg px-3 py-1.5 text-xs font-semibold text-white transition-colors', platformButtonClass(subscription.group?.platform || '')]"
                @click="router.push({ path: '/purchase', query: { tab: 'subscription', group: String(subscription.group_id) } })"
              >
                {{ t('payment.renewNow') }}
              </button>
            </div>
          </div>

          <!-- Usage Progress -->
          <div class="space-y-4 p-4">
            <!-- Expiration Info -->
            <div v-if="subscription.expires_at" class="flex items-center justify-between text-sm">
              <span class="text-gray-500 dark:text-dark-400">{{
                t('userSubscriptions.expires')
              }}</span>
              <span :class="getExpirationClass(subscription.expires_at)">
                {{ formatExpirationDate(subscription.expires_at) }}
              </span>
            </div>
            <div v-else class="flex items-center justify-between text-sm">
              <span class="text-gray-500 dark:text-dark-400">{{
                t('userSubscriptions.expires')
              }}</span>
              <span class="text-gray-700 dark:text-gray-300">{{
                t('userSubscriptions.noExpiration')
              }}</span>
            </div>

            <!-- Daily Usage -->
            <div v-if="subscription.group?.daily_limit_usd" class="space-y-2">
              <div class="flex items-center justify-between">
                <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
                  {{ t('userSubscriptions.daily') }}
                </span>
                <span class="text-sm text-gray-500 dark:text-dark-400">
                  ${{ (subscription.daily_usage_usd || 0).toFixed(2) }} / ${{
                    subscription.group.daily_limit_usd.toFixed(2)
                  }}
                </span>
              </div>
              <div class="relative h-2 overflow-hidden rounded-full bg-gray-200 dark:bg-dark-600">
                <div
                  class="absolute inset-y-0 left-0 rounded-full transition-all duration-300"
                  :class="
                    getProgressBarClass(
                      subscription.daily_usage_usd,
                      subscription.group.daily_limit_usd
                    )
                  "
                  :style="{
                    width: getProgressWidth(
                      subscription.daily_usage_usd,
                      subscription.group.daily_limit_usd
                    )
                  }"
                ></div>
              </div>
              <p
                v-if="subscription.daily_window_start"
                class="text-xs text-gray-500 dark:text-dark-400"
              >
                {{ formatDailyUsageWindow(subscription) }}
              </p>
            </div>

            <!-- Weekly Usage -->
            <div v-if="subscription.group?.weekly_limit_usd" class="space-y-2">
              <div class="flex items-center justify-between">
                <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
                  {{ t('userSubscriptions.weekly') }}
                </span>
                <span class="text-sm text-gray-500 dark:text-dark-400">
                  ${{ (subscription.weekly_usage_usd || 0).toFixed(2) }} / ${{
                    subscription.group.weekly_limit_usd.toFixed(2)
                  }}
                </span>
              </div>
              <div class="relative h-2 overflow-hidden rounded-full bg-gray-200 dark:bg-dark-600">
                <div
                  class="absolute inset-y-0 left-0 rounded-full transition-all duration-300"
                  :class="
                    getProgressBarClass(
                      subscription.weekly_usage_usd,
                      subscription.group.weekly_limit_usd
                    )
                  "
                  :style="{
                    width: getProgressWidth(
                      subscription.weekly_usage_usd,
                      subscription.group.weekly_limit_usd
                    )
                  }"
                ></div>
              </div>
              <p
                v-if="subscription.weekly_window_start"
                class="text-xs text-gray-500 dark:text-dark-400"
              >
                {{
                  t('userSubscriptions.resetIn', {
                    time: formatResetTime(subscription.weekly_window_start, 168)
                  })
                }}
              </p>
            </div>

            <!-- Monthly Usage -->
            <div v-if="subscription.group?.monthly_limit_usd" class="space-y-2">
              <div class="flex items-center justify-between">
                <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
                  {{ t('userSubscriptions.monthly') }}
                </span>
                <span class="text-sm text-gray-500 dark:text-dark-400">
                  ${{ (subscription.monthly_usage_usd || 0).toFixed(2) }} / ${{
                    subscription.group.monthly_limit_usd.toFixed(2)
                  }}
                </span>
              </div>
              <div class="relative h-2 overflow-hidden rounded-full bg-gray-200 dark:bg-dark-600">
                <div
                  class="absolute inset-y-0 left-0 rounded-full transition-all duration-300"
                  :class="
                    getProgressBarClass(
                      subscription.monthly_usage_usd,
                      subscription.group.monthly_limit_usd
                    )
                  "
                  :style="{
                    width: getProgressWidth(
                      subscription.monthly_usage_usd,
                      subscription.group.monthly_limit_usd
                    )
                  }"
                ></div>
              </div>
              <p
                v-if="subscription.monthly_window_start"
                class="text-xs text-gray-500 dark:text-dark-400"
              >
                {{
                  t('userSubscriptions.resetIn', {
                    time: formatResetTime(subscription.monthly_window_start, 720)
                  })
                }}
              </p>
            </div>

            <!-- No limits configured - Unlimited badge -->
            <div
              v-if="
                !subscription.group?.daily_limit_usd &&
                !subscription.group?.weekly_limit_usd &&
                !subscription.group?.monthly_limit_usd &&
                !officialUsageFor(subscription)
              "
              data-testid="unlimited-box"
              class="flex items-center justify-center rounded-xl bg-gradient-to-r from-emerald-50 to-teal-50 py-6 dark:from-emerald-900/20 dark:to-teal-900/20"
            >
              <div class="flex items-center gap-3">
                <span class="text-4xl text-emerald-600 dark:text-emerald-400">∞</span>
                <div>
                  <p class="text-sm font-medium text-emerald-700 dark:text-emerald-300">
                    {{ t('userSubscriptions.unlimited') }}
                  </p>
                  <p class="text-xs text-emerald-600/70 dark:text-emerald-400/70">
                    {{ t('userSubscriptions.unlimitedDesc') }}
                  </p>
                </div>
              </div>
            </div>

            <!-- Official upstream subscription usage (same logic as /quota) -->
            <div
              v-if="showOfficialUsageSection(subscription)"
              :data-testid="`official-usage-${subscription.group_id}`"
              class="space-y-3 rounded-xl border border-gray-100 p-3 dark:border-dark-700"
            >
              <div class="flex items-center justify-between gap-2">
                <div class="flex items-center gap-2">
                  <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
                    {{ t('userSubscriptions.officialUsage.title') }}
                  </span>
                  <span
                    v-if="officialUsageFor(subscription)?.plan"
                    class="rounded-md bg-gray-100 px-1.5 py-0.5 text-[11px] font-medium uppercase text-gray-600 dark:bg-dark-700 dark:text-gray-300"
                  >
                    {{ officialUsageFor(subscription)?.plan }}
                  </span>
                </div>
                <span
                  v-if="officialUsageFor(subscription)?.updated_at"
                  class="text-[11px] text-gray-400 dark:text-gray-500"
                >
                  {{ t('userSubscriptions.officialUsage.updatedAt', { time: formatClock(officialUsageFor(subscription)?.updated_at) }) }}
                </span>
              </div>

              <div
                v-if="usageLoading && !usageByGroup.has(subscription.group_id)"
                :data-testid="`official-usage-loading-${subscription.group_id}`"
                class="flex items-center gap-2 text-xs text-gray-500 dark:text-dark-400"
              >
                <div class="h-3 w-3 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"></div>
                {{ t('userSubscriptions.officialUsage.loading') }}
              </div>

              <p
                v-else-if="officialUsageErrorText(subscription)"
                class="text-xs text-gray-500 dark:text-dark-400"
              >
                {{ officialUsageErrorText(subscription) }}
              </p>

              <template v-else-if="officialUsageFor(subscription)">
                <div
                  v-for="win in officialUsageWindows(officialUsageFor(subscription))"
                  :key="win.key"
                  class="space-y-1"
                >
                  <div class="flex items-center justify-between text-xs">
                    <span class="font-medium text-gray-700 dark:text-gray-300">{{ t(win.labelKey) }}</span>
                    <span class="text-gray-500 dark:text-dark-400">
                      {{ officialPercent(win.window.utilization) }}%<template v-if="win.window.resets_at"> · {{ t('userSubscriptions.officialUsage.resetsAt', { time: formatResetsAt(win.window.resets_at) }) }}</template>
                    </span>
                  </div>
                  <div class="relative h-2 overflow-hidden rounded-full bg-gray-200 dark:bg-dark-600">
                    <div
                      data-testid="official-usage-bar"
                      class="absolute inset-y-0 left-0 rounded-full transition-all duration-300"
                      :class="getProgressBarClass(clampPercent(win.window.utilization), 100)"
                      :style="{ width: `${clampPercent(win.window.utilization)}%` }"
                    ></div>
                  </div>
                </div>
                <p
                  v-if="officialUsageWindows(officialUsageFor(subscription)).length === 0"
                  class="text-xs text-gray-500 dark:text-dark-400"
                >
                  {{ t('userSubscriptions.officialUsage.unavailable') }}
                </p>
              </template>
            </div>

            <!-- Current 30-day period usage valued at official API prices -->
            <div
              v-if="periodUsageFor(subscription)"
              :data-testid="`period-usage-${subscription.id}`"
              class="space-y-1 rounded-xl border border-gray-100 p-3 dark:border-dark-700"
            >
              <div class="flex flex-wrap items-baseline justify-between gap-x-2 gap-y-0.5">
                <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
                  {{ t('userSubscriptions.periodUsage.title') }}
                </span>
                <span class="text-[11px] text-gray-400 dark:text-gray-500">
                  {{ periodSubtitle(periodUsageFor(subscription)!) }}
                </span>
              </div>
              <p class="text-lg font-semibold text-gray-900 dark:text-white">
                {{ t('userSubscriptions.periodUsage.amount', { amount: formatUsd(periodUsageFor(subscription)!.total_cost_usd) }) }}
              </p>
              <p class="text-xs text-gray-500 dark:text-dark-400">
                {{ periodSummary(periodUsageFor(subscription)!) }}
              </p>
            </div>
            <div
              v-else-if="isPeriodUsageEligible(subscription) && periodLoading"
              :data-testid="`period-usage-loading-${subscription.id}`"
              class="flex items-center gap-2 text-xs text-gray-500 dark:text-dark-400"
            >
              <div class="h-3 w-3 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"></div>
              {{ t('userSubscriptions.periodUsage.loading') }}
            </div>
          </div>
        </div>
      </div>

      <!-- Official usage: whole-request failure notice + refresh -->
      <div
        v-if="!loading && hasActiveSubscriptions"
        class="flex flex-wrap items-center justify-end gap-3 text-xs text-gray-500 dark:text-dark-400"
      >
        <span v-if="usageFailed" data-testid="official-usage-notice">
          {{ t('userSubscriptions.officialUsage.requestFailed') }}
        </span>
        <span v-if="periodFailed" data-testid="period-usage-notice">
          {{ t('userSubscriptions.periodUsage.requestFailed') }}
        </span>
        <button
          type="button"
          data-testid="official-usage-refresh"
          class="rounded-lg border border-gray-200 px-2.5 py-1 font-medium text-gray-600 transition-colors hover:bg-gray-50 disabled:opacity-50 dark:border-dark-600 dark:text-gray-300 dark:hover:bg-dark-700"
          :disabled="usageLoading || periodLoading"
          @click="refreshUsage"
        >
          {{ t('userSubscriptions.refreshUsage') }}
        </button>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { useAppStore } from '@/stores/app'
import subscriptionsAPI from '@/api/subscriptions'
import type {
  UserSubscription,
  OfficialSubscriptionUsage,
  OfficialUsageWindow,
  SubscriptionUsageItem,
  SubscriptionPeriodUsageItem
} from '@/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { formatCompactNumber, formatDateTimeToMinute } from '@/utils/format'
import { hasPeakRate, formatPeakRateWindow, serverTimezoneLabel } from '@/utils/peak-rate'
import { platformBorderClass, platformBadgeClass, platformButtonClass, platformLabel } from '@/utils/platformColors'
import {
  getExpirationDateRelation,
  getRemainingDurationParts,
  isOneTimeDailyQuota,
  type RemainingDurationParts
} from '@/utils/subscriptionQuota'

function platformAccentDotClass(p: string): string {
  switch (p) {
    case 'anthropic': return 'bg-orange-500'
    case 'openai': return 'bg-emerald-500'
    case 'antigravity': return 'bg-purple-500'
    case 'gemini': return 'bg-blue-500'
    default: return 'bg-gray-400'
  }
}

const { t } = useI18n()
const router = useRouter()
const appStore = useAppStore()

const subscriptions = ref<UserSubscription[]>([])
const loading = ref(true)

function subscriptionHasPeakRate(subscription: UserSubscription): boolean {
  return hasPeakRate(subscription.group)
}

function subscriptionPeakRateLabel(subscription: UserSubscription): string {
  return formatPeakRateWindow(subscription.group, serverTimezoneLabel(appStore.cachedPublicSettings?.server_utc_offset))
}

async function loadSubscriptions() {
  try {
    loading.value = true
    subscriptions.value = await subscriptionsAPI.getMySubscriptions()
  } catch (error) {
    console.error('Failed to load subscriptions:', error)
    appStore.showError(t('userSubscriptions.failedToLoad'))
  } finally {
    loading.value = false
  }
}

// ==================== Official upstream subscription usage ====================

const usageByGroup = ref(new Map<number, SubscriptionUsageItem>())
const usageLoading = ref(false)
const usageFailed = ref(false)

const hasOfficialUsageSubscriptions = computed(() => subscriptions.value.some(isOfficialUsageEligible))

const OFFICIAL_WINDOWS = [
  { key: 'five_hour', labelKey: 'userSubscriptions.officialUsage.fiveHour' },
  { key: 'seven_day', labelKey: 'userSubscriptions.officialUsage.sevenDay' },
  { key: 'seven_day_sonnet', labelKey: 'userSubscriptions.officialUsage.sevenDaySonnet' },
  { key: 'seven_day_fable', labelKey: 'userSubscriptions.officialUsage.sevenDayFable' }
] as const

async function loadOfficialUsage() {
  if (usageLoading.value) return
  usageFailed.value = false
  usageLoading.value = true
  try {
    const items = await subscriptionsAPI.getSubscriptionUsage()
    const next = new Map<number, SubscriptionUsageItem>()
    for (const item of items || []) next.set(item.group_id, item)
    usageByGroup.value = next
  } catch (error) {
    console.error('Failed to load official subscription usage:', error)
    usageFailed.value = true
  } finally {
    usageLoading.value = false
  }
}

/** Official upstream usage only applies to Anthropic subscription groups. */
function isOfficialUsageEligible(subscription: UserSubscription): boolean {
  return subscription.status === 'active' && subscription.group?.platform === 'anthropic'
}

function showOfficialUsageSection(subscription: UserSubscription): boolean {
  if (!isOfficialUsageEligible(subscription)) return false
  if (usageByGroup.value.has(subscription.group_id)) return true
  return usageLoading.value && !usageFailed.value
}

function officialUsageFor(subscription: UserSubscription): OfficialSubscriptionUsage | null {
  if (!isOfficialUsageEligible(subscription)) return null
  return usageByGroup.value.get(subscription.group_id)?.usage ?? null
}

function officialUsageErrorText(subscription: UserSubscription): string {
  if (!isOfficialUsageEligible(subscription)) return ''
  const item = usageByGroup.value.get(subscription.group_id)
  if (!item || item.usage) return ''
  switch (item.error?.type) {
    case 'not_subscription_group':
      return t('userSubscriptions.officialUsage.provisioning')
    case 'no_available_account':
      return t('userSubscriptions.officialUsage.noAvailableAccount')
    default:
      return t('userSubscriptions.officialUsage.unavailable')
  }
}

function officialUsageWindows(usage: OfficialSubscriptionUsage | null) {
  if (!usage) return []
  const out: Array<{ key: string; labelKey: string; window: OfficialUsageWindow }> = []
  for (const w of OFFICIAL_WINDOWS) {
    const window = usage[w.key]
    if (window) out.push({ key: w.key, labelKey: w.labelKey, window })
  }
  return out
}

function clampPercent(utilization: number | null | undefined): number {
  const v = Number(utilization)
  if (!Number.isFinite(v)) return 0
  return Math.min(Math.max(v, 0), 100)
}

/** Real rounded utilization like /quota (may exceed 100); only the bar width is clamped. */
function officialPercent(utilization: number | null | undefined): number {
  const v = Number(utilization)
  return Number.isFinite(v) ? Math.round(v) : 0
}

const pad2 = (n: number) => String(n).padStart(2, '0')

function formatClock(value: string | null | undefined): string {
  if (!value) return ''
  const d = new Date(value)
  if (Number.isNaN(d.getTime())) return ''
  return `${pad2(d.getHours())}:${pad2(d.getMinutes())}`
}

/** Mirrors the /quota command: "HH:mm" when the reset is today (local), otherwise "MM-DD HH:mm". */
function formatResetsAt(value: string | null | undefined): string {
  if (!value) return ''
  const d = new Date(value)
  if (Number.isNaN(d.getTime())) return ''
  const now = new Date()
  const sameDay =
    d.getFullYear() === now.getFullYear() && d.getMonth() === now.getMonth() && d.getDate() === now.getDate()
  const clock = `${pad2(d.getHours())}:${pad2(d.getMinutes())}`
  return sameDay ? clock : `${pad2(d.getMonth() + 1)}-${pad2(d.getDate())} ${clock}`
}

// ==================== Current 30-day period usage (official API prices) ====================

const periodBySubscription = ref(new Map<number, SubscriptionPeriodUsageItem>())
const periodLoading = ref(false)
const periodFailed = ref(false)

const hasActiveSubscriptions = computed(() => subscriptions.value.some(isPeriodUsageEligible))

function isPeriodUsageEligible(subscription: UserSubscription): boolean {
  return subscription.status === 'active'
}

async function loadPeriodUsage() {
  if (periodLoading.value) return
  periodFailed.value = false
  periodLoading.value = true
  try {
    const items = await subscriptionsAPI.getSubscriptionPeriodUsage()
    const next = new Map<number, SubscriptionPeriodUsageItem>()
    for (const item of items || []) next.set(item.subscription_id, item)
    periodBySubscription.value = next
  } catch (error) {
    console.error('Failed to load subscription period usage:', error)
    periodFailed.value = true
  } finally {
    periodLoading.value = false
  }
}

/** Loads official usage (Anthropic only) and period usage (all active) in parallel. */
async function refreshUsage() {
  const tasks: Promise<void>[] = []
  if (hasOfficialUsageSubscriptions.value) tasks.push(loadOfficialUsage())
  if (hasActiveSubscriptions.value) tasks.push(loadPeriodUsage())
  await Promise.all(tasks)
}

function periodUsageFor(subscription: UserSubscription): SubscriptionPeriodUsageItem | null {
  if (!isPeriodUsageEligible(subscription)) return null
  return periodBySubscription.value.get(subscription.id) ?? null
}

const usdFormatter = new Intl.NumberFormat('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })

function formatUsd(amount: number | null | undefined): string {
  const v = Number(amount)
  return `$${usdFormatter.format(Number.isFinite(v) ? v : 0)}`
}

const DAY_MS = 24 * 60 * 60 * 1000

function formatMonthDay(value: string | number): string {
  const d = new Date(value)
  if (Number.isNaN(d.getTime())) return ''
  return `${pad2(d.getMonth() + 1)}-${pad2(d.getDate())}`
}

/** "Period N · MM-DD ~ MM-DD" in local dates; the end shows the inclusive last day (period_end - 1 day). */
function periodSubtitle(item: SubscriptionPeriodUsageItem): string {
  const end = new Date(item.period_end).getTime() - DAY_MS
  return t('userSubscriptions.periodUsage.period', {
    index: item.period_index,
    start: formatMonthDay(item.period_start),
    end: formatMonthDay(end)
  })
}

function periodSummary(item: SubscriptionPeriodUsageItem): string {
  if (!item.request_count) return t('userSubscriptions.periodUsage.empty')
  return t('userSubscriptions.periodUsage.summary', {
    requests: item.request_count,
    input: formatCompactNumber(item.input_tokens),
    output: formatCompactNumber(item.output_tokens),
    cacheWrite: formatCompactNumber(item.cache_creation_tokens),
    cacheRead: formatCompactNumber(item.cache_read_tokens)
  })
}

function getProgressWidth(used: number | undefined, limit: number | null | undefined): string {
  if (!limit || limit === 0) return '0%'
  const percentage = Math.min(((used || 0) / limit) * 100, 100)
  return `${percentage}%`
}

function getProgressBarClass(used: number | undefined, limit: number | null | undefined): string {
  if (!limit || limit === 0) return 'bg-gray-400'
  const percentage = ((used || 0) / limit) * 100
  if (percentage >= 90) return 'bg-red-500'
  if (percentage >= 70) return 'bg-orange-500'
  return 'bg-green-500'
}

function formatExpirationDate(expiresAt: string): string {
  const now = new Date()
  const expires = new Date(expiresAt)
  const diff = expires.getTime() - now.getTime()
  const days = Math.ceil(diff / (1000 * 60 * 60 * 24))
  const relation = getExpirationDateRelation(expires, now)

  if (relation === null) return ''

  if (relation === 'expired') {
    return t('userSubscriptions.status.expired')
  }

  const dateStr = formatDateTimeToMinute(expires)

  if (relation === 'today') {
    return `${dateStr} (${t('common.today')})`
  }
  if (relation === 'tomorrow') {
    return `${dateStr} (${t('common.tomorrow')})`
  }

  return t('userSubscriptions.daysRemaining', { days }) + ` (${dateStr})`
}

function getExpirationClass(expiresAt: string): string {
  const now = new Date()
  const expires = new Date(expiresAt)
  const diff = expires.getTime() - now.getTime()
  const days = Math.ceil(diff / (1000 * 60 * 60 * 24))

  if (diff <= 0) return 'text-red-600 dark:text-red-400 font-medium'
  if (days <= 3) return 'text-red-600 dark:text-red-400'
  if (days <= 7) return 'text-orange-600 dark:text-orange-400'
  return 'text-gray-700 dark:text-gray-300'
}

function formatDurationParts(parts: RemainingDurationParts): string {
  if (parts.days > 0) {
    return `${parts.days}d ${parts.hours}h`
  }

  if (parts.hours > 0) {
    return `${parts.hours}h ${parts.minutes}m`
  }

  return `${parts.minutes}m`
}

function formatDailyUsageWindow(subscription: UserSubscription): string {
  if (isOneTimeDailyQuota(subscription) && subscription.expires_at) {
    const parts = getRemainingDurationParts(subscription.expires_at)
    if (!parts) return t('userSubscriptions.windowNotActive')
    return t('userSubscriptions.quotaEndsIn', { time: formatDurationParts(parts) })
  }

  return t('userSubscriptions.resetIn', {
    time: formatResetTime(subscription.daily_window_start, 24)
  })
}

function formatResetTime(windowStart: string | null, windowHours: number): string {
  if (!windowStart) return t('userSubscriptions.windowNotActive')

  const start = new Date(windowStart)
  const end = new Date(start.getTime() + windowHours * 60 * 60 * 1000)
  const parts = getRemainingDurationParts(end)

  return parts ? formatDurationParts(parts) : t('userSubscriptions.windowNotActive')
}

onMounted(async () => {
  await loadSubscriptions()
  void refreshUsage()
})
</script>
