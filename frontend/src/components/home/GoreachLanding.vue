<template>
  <div ref="rootEl" class="goreach" @click="onAnchorClick">
    <header class="nav">
      <div class="wrap">
        <router-link class="brand" to="/home" aria-label="Goreach 首页"
          ><img v-if="siteLogo" class="brand-logo" :src="siteLogo" alt="" aria-hidden="true" /><span
            v-else
            class="brand-mark"
            aria-hidden="true"
          ></span
          >Goreach</router-link
        >
        <nav class="nav-links" aria-label="页面导航">
          <a href="#compare">风险对比</a>
          <a href="#pricing">价格</a>
          <a href="#how">使用方式</a>
          <a href="#faq">常见问题</a>
        </nav>
        <div class="nav-auth">
          <router-link v-if="isAuthenticated" class="gr-btn gr-btn-primary gr-btn-sm" :to="dashboardPath"
            >进入控制台</router-link
          >
          <template v-else>
            <router-link class="gr-btn gr-btn-plain gr-btn-sm" to="/login">登录</router-link>
            <router-link class="gr-btn gr-btn-primary gr-btn-sm" to="/register">注册</router-link>
          </template>
        </div>
      </div>
    </header>

    <main>
      <section class="hero">
        <div class="wrap hero-grid">
          <div>
            <span class="eyebrow">100% 官方订阅 · 托管账号环境</span>
            <h1>官方账号，从开通到稳定使用，由 <span class="latin">Goreach</span> 托管</h1>
            <p class="lede">
              境外主体注册、企业商业卡统一支付、办公室固定出口 IP。用量可在官方后台核验，代码与对话零留存。
            </p>
            <div class="hero-cta">
              <a v-if="isAuthenticated" class="gr-btn gr-btn-primary" href="#pricing">查看档位</a>
              <router-link v-else class="gr-btn gr-btn-primary" to="/register">注册并选择档位</router-link>
              <a class="gr-btn gr-btn-ghost" href="#compare">先看风险对比</a>
              <span v-if="isAuthenticated" class="signin"
                >已登录，<router-link :to="dashboardPath">进入控制台</router-link></span
              >
              <span v-else class="signin">已有账号？<router-link to="/login">登录</router-link></span>
            </div>
          </div>

          <div class="seat-wrap" aria-label="Goreach 托管席位包含的内容">
            <div class="seat-card">
              <div class="seat-head">
                <span class="seat-issuer">Goreach · 托管席位</span>
                <span class="seat-status">官方订阅</span>
              </div>
              <p class="seat-plan">Claude Enterprise<small>官方企业版订阅席位，非按量计费中转</small></p>
              <div class="seat-perf" aria-hidden="true"></div>
              <dl class="seat-rows">
                <div><dt>席位</dt><dd>独立席位，一人一号</dd></div>
                <div><dt>出口</dt><dd>办公室固定 IP，不与他人共用</dd></div>
                <div><dt>支付</dt><dd>境外主体 + 企业商业卡</dd></div>
                <div><dt>用量</dt><dd>可在官方后台直接核验</dd></div>
                <div><dt>留存</dt><dd>不接触、不留存业务数据</dd></div>
              </dl>
              <div class="seat-foot">
                <span class="seat-no">可选档位<b>Standard / Premium</b></span>
                <span class="barcode" aria-hidden="true"></span>
              </div>
            </div>
            <p class="seat-caption">每个订阅对应一个独立的官方席位</p>
          </div>
        </div>

        <div class="wrap">
          <p class="notice" role="note">
            <b>开通提示</b
            ><span>近期上游风控趋严，账号开通流程明显变慢，新订单预计一周以上才能开通；已有订阅不受影响。</span>
          </p>
        </div>
      </section>

      <section id="compare" class="block">
        <div class="wrap">
          <p class="kicker">风险与对比</p>
          <h2>为什么自己开的号容易出问题</h2>
          <p class="section-lede">
            把常见的账号异常诱因逐条摊开对照。下表为方案说明，具体以官方政策与实际服务条款为准。
          </p>

          <div class="policy">
            <span class="policy-tag">须知</span>
            <div>
              <h3>上游政策风险</h3>
              <p>
                本方案依赖上游厂商的地区与账号政策。2026 年 11 月 12
                日生效的地区与实控穿透规则可能直接影响可用性，此类变化不受服务方控制。付费前请确认你已理解这一点。
              </p>
            </div>
          </div>

          <table class="compare">
            <thead>
              <tr>
                <th scope="col">风险点</th>
                <th scope="col">自行注册</th>
                <th scope="col">普通中转 / 合租</th>
                <th scope="col" class="ours">Goreach 方案</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="row in compareRows" :key="row.risk">
                <th scope="row">{{ row.risk }}</th>
                <td data-label="自行注册">{{ row.self }}</td>
                <td data-label="普通中转">{{ row.relay }}</td>
                <td class="ours">{{ row.ours }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      <section id="pricing" class="block">
        <div class="wrap">
          <p class="kicker">价格</p>
          <h2>两档月订阅，都是官方企业版席位</h2>
          <p class="plans-intro">
            Goreach 提供的是 Claude 官方企业版账号，按 Anthropic 官方定价：<b
              >Standard 档用量约为 Pro 的 1.25 倍，Premium 档约为 6.25 倍</b
            >。不是按量计费的 API 中转。
          </p>

          <div class="plans">
            <article class="plan">
              <h3>Standard</h3>
              <p class="ratio">约 1.25× Pro 用量</p>
              <div class="price"><span class="cur">$</span><span class="num">45</span><span class="per">/ 月</span></div>
              <p class="was">原价 $60 / 月</p>
              <ul>
                <li>适合日常单项目，常规调用频次</li>
              </ul>
              <router-link
                class="gr-btn gr-btn-ghost"
                :to="isAuthenticated ? dashboardPath : '/register?plan=standard'"
                >{{ isAuthenticated ? '选择 Standard' : '注册并选择 Standard' }}</router-link
              >
            </article>
            <article class="plan featured">
              <span class="plan-badge">推荐</span>
              <h3>Premium</h3>
              <p class="ratio">约 6.25× Pro 用量</p>
              <div class="price"><span class="cur">$</span><span class="num">180</span><span class="per">/ 月</span></div>
              <p class="was">原价 $220 / 月</p>
              <ul>
                <li>适合高频调用，多项目并行；优先开通，优先异常协助排查</li>
              </ul>
              <router-link
                class="gr-btn gr-btn-primary"
                :to="isAuthenticated ? dashboardPath : '/register?plan=premium'"
                >{{ isAuthenticated ? '选择 Premium' : '注册并选择 Premium' }}</router-link
              >
            </article>
          </div>
          <p class="plans-note">
            按月订阅，可随时取消。退款按服务条款执行，退款将扣除订单金额的 25% 作为已发生成本。
          </p>
        </div>
      </section>

      <section id="how" class="block">
        <div class="wrap">
          <p class="kicker">使用方式</p>
          <h2>四步开始使用</h2>
          <p class="section-lede">不需要在机器上安装额外的代理软件或导入配置文件，也无需切换 AI 工作台。</p>
          <ol class="steps">
            <li>
              <h3>选择档位并支付</h3>
              <p>下单后完成支付，支持境内银行卡、支付宝和微信支付，按交易时汇率结算人民币。</p>
            </li>
            <li>
              <h3>等待账号开通</h3>
              <p>由我们完成账号注册、支付绑定与环境配置；风控趋严时预计一周以上。</p>
            </li>
            <li>
              <h3>登录 Goreach</h3>
              <p>通过系统代理完成登录，查看订阅信息、模型用量监控与模型基础配置。</p>
            </li>
            <li>
              <h3>在常用工具里调用</h3>
              <p>直接接入你正在用的工具，无需切换工作台。</p>
              <div class="tools">
                <span>Claude Code</span><span>Claude 桌面端</span><span>WorkBuddy</span><span>CC Switch</span>
              </div>
            </li>
          </ol>
        </div>
      </section>

      <section id="faq" class="block">
        <div class="wrap">
          <p class="kicker">常见疑问</p>
          <h2>先回答最常被问到的几个</h2>
          <p class="section-lede">关于账号、支付、退款与开通。</p>
          <div class="faq">
            <details v-for="(item, index) in faqs" :key="item.q" :open="index === 0">
              <summary>{{ item.q }}</summary>
              <p>{{ item.a }}</p>
            </details>
          </div>
        </div>
      </section>

      <section id="reading" class="block">
        <div class="wrap">
          <p class="kicker">延伸阅读</p>
          <h2>做决定之前，建议一起看完</h2>
          <p class="section-lede">政策、定价、合规与条款的原始材料。</p>
          <div class="reading">
            <a
              v-for="item in readings"
              :key="item.title"
              :href="item.href"
              :target="item.external ? '_blank' : undefined"
              :rel="item.external ? 'noopener' : undefined"
              ><span class="tag">{{ item.tag }}</span><b>{{ item.title }}</b
              ><span class="desc">{{ item.desc }}</span
              ><span v-if="item.todo" class="todo">{{ item.todo }}</span></a
            >
          </div>
        </div>
      </section>

      <section class="closing">
        <div class="wrap">
          <div class="box">
            <div>
              <h2>准备好了就注册，选好档位再付款</h2>
              <p>开通进度、订阅信息和模型用量，都在 Goreach 控制台里查看。</p>
            </div>
            <div class="hero-cta">
              <router-link v-if="isAuthenticated" class="gr-btn gr-btn-primary" :to="dashboardPath"
                >进入控制台</router-link
              >
              <template v-else>
                <router-link class="gr-btn gr-btn-primary" to="/register">注册</router-link>
                <router-link class="gr-btn gr-btn-ghost" to="/login">登录</router-link>
              </template>
            </div>
          </div>
        </div>
      </section>
    </main>

    <footer>
      <div class="wrap">
        <span>© 2026 Goreach</span>
        <span
          ><a href="#reading">服务条款</a
          ><router-link v-if="isAuthenticated" :to="dashboardPath">进入控制台</router-link
          ><template v-else
            ><router-link to="/login">登录</router-link><router-link to="/register">注册</router-link></template
          ></span
        >
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

withDefaults(
  defineProps<{
    isAuthenticated: boolean
    dashboardPath: string
    /** 后台「站点 Logo」（已经过 sanitizeUrl）；为空时显示内置的简易标记 */
    siteLogo?: string
  }>(),
  { siteLogo: '' },
)

const rootEl = ref<HTMLElement | null>(null)

const compareRows = [
  { risk: '支付地区与账单信息', self: '常与账号地区不一致', relay: '共用支付渠道，来源混杂', ours: '境外主体 + 企业商业卡支付' },
  { risk: '出口 IP 稳定性', self: '机场 / 代理，频繁变动', relay: '共享出口，特征高度重叠', ours: '办公室固定出口，不与他人共用' },
  { risk: '账号归属与共享特征', self: '个人号，风控敏感', relay: '多人共用，最易被识别', ours: '独立席位，一人一号' },
  { risk: '客户端与模型来源', self: '官方客户端', relay: '镜像接口，来源不可核验', ours: '100% 官方客户端与官方模型' },
  { risk: '用量是否可核验', self: '可核验', relay: '仅平台自有后台', ours: '可在官方后台直接核对' },
  { risk: '代码与对话留存', self: '按官方政策', relay: '存在留存风险', ours: '不接触、不留存业务数据' },
]

const faqs = [
  {
    q: '和 API 中转站有什么区别？',
    a: '我们提供的是官方订阅席位与官方客户端，不是按量计费的镜像接口。用量可以在官方后台直接核验，我们不接触、不留存你的代码与对话内容。',
  },
  {
    q: '需要自己准备代理吗？',
    a: '登录步骤经由你的系统代理完成，此后的流量通过办公室固定出口。除此之外不需要安装额外软件或导入配置文件。',
  },
  {
    q: '为什么开通要等一周以上？',
    a: '近期上游风控趋严，主体注册、支付绑定与环境配置需要逐个完成，无法批量秒开。已有订阅不受影响。',
  },
  {
    q: '支持哪些支付方式？',
    a: '支持境内银行卡、支付宝与微信支付，按交易发生时的汇率结算人民币，不占用个人购汇额度。商户为境外主体，收单由持牌支付机构完成。',
  },
  {
    q: '退款怎么算？',
    a: '按月订阅可随时取消。退款按服务条款执行，将扣除订单金额的 25% 作为已发生成本，其余部分原路退回。',
  },
  {
    q: '账号被风控或政策变化怎么办？',
    a: '我们协助排查账号环境问题，但上游厂商的地区与账号政策不受我们控制，由此导致的可用性变化无法承诺补偿。付费前请阅读下方延伸阅读中的政策说明。',
  },
]

const readings = [
  {
    href: 'https://support.anthropic.com/',
    external: true,
    tag: '政策',
    title: '上游地区与账号政策说明（2026.11.12 生效）',
    desc: '使用者物理位置、实控穿透与转售限制的关键条款，直接影响服务可用性。',
    todo: '链接待替换为具体条款页',
  },
  {
    href: 'https://www.anthropic.com/pricing',
    external: true,
    tag: '定价',
    title: '官方订阅档位与用量规格',
    desc: 'Pro / Team / Max 各档用量与价格的官方说明，可用于核对席位规格。',
    todo: '',
  },
  {
    href: '#',
    external: false,
    tag: '自查',
    title: '账号环境风险自查清单',
    desc: '支付地区、出口 IP、共享特征、客户端来源四项对照，判断自己是否适合托管方案。',
    todo: '链接待补',
  },
  {
    href: '#',
    external: false,
    tag: '合规',
    title: '生成式 AI 服务与跨境使用提示',
    desc: '境内提供生成式 AI 服务的备案要求、国际联网相关规定与数据出境注意事项。',
    todo: '链接待补',
  },
  {
    href: '#',
    external: false,
    tag: '条款',
    title: '服务条款与退款细则',
    desc: '订阅周期、取消方式、退款扣费比例与争议处理方式的完整文本。',
    todo: '链接待补',
  },
  {
    href: '#',
    external: false,
    tag: '替代',
    title: '合规渠道与替代方案对比',
    desc: '云厂商官方 API 渠道、国产模型工具链的成本与体验对照。',
    todo: '链接待补',
  },
]

/**
 * In-page anchors (#compare, #pricing ...) would otherwise change location.hash,
 * which Vue Router (web history) picks up via popstate and then applies the
 * global scrollBehavior ({ top: 0 }), jumping back to the top. Handle them here.
 */
function onAnchorClick(event: MouseEvent) {
  if (event.defaultPrevented || event.button !== 0 || event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) {
    return
  }
  const anchor = (event.target as Element | null)?.closest?.('a[href^="#"]')
  if (!anchor) return
  event.preventDefault()
  const id = anchor.getAttribute('href')?.slice(1) ?? ''
  const target = id ? document.getElementById(id) : rootEl.value
  if (!target) return
  const reduceMotion =
    typeof window.matchMedia === 'function' && window.matchMedia('(prefers-reduced-motion: reduce)').matches
  target.scrollIntoView({ behavior: reduceMotion ? 'auto' : 'smooth', block: 'start' })
}
</script>

<style scoped>
.goreach {
  --ink: #14162b;
  --ink-soft: #4a4e68;
  --ink-faint: #8a8ea6;
  --paper: #f6f7fb;
  --card: #ffffff;
  --line: #e3e5ef;
  --seat: #3d4bd8;
  --seat-deep: #2a35a8;
  --seat-wash: #eceefc;
  --verified: #12936b;
  --verified-wash: #e7f5ef;
  --caution: #9a6412;
  --caution-wash: #fbf3e4;
  --danger: #b4322e;
  --radius: 14px;
  --sans: 'PingFang SC', 'HarmonyOS Sans SC', 'Microsoft YaHei', 'Noto Sans CJK SC', system-ui, -apple-system,
    'Segoe UI', sans-serif;
  --latin: 'Avenir Next', 'Segoe UI', 'Helvetica Neue', var(--sans);
  --mono: ui-monospace, 'SF Mono', Menlo, Consolas, monospace;

  /* self-contained light theme, independent of the app's dark mode / Tailwind base */
  color-scheme: light;
  min-height: 100vh;
  margin: 0;
  background: var(--paper);
  color: var(--ink);
  font: 16px/1.7 var(--sans);
  -webkit-font-smoothing: antialiased;
}
.goreach *,
.goreach *::before,
.goreach *::after {
  box-sizing: border-box;
}
a {
  color: inherit;
}
/* section anchors land below the sticky nav */
section[id] {
  scroll-margin-top: 64px;
}
:focus-visible {
  outline: 3px solid var(--seat);
  outline-offset: 3px;
  border-radius: 6px;
}
.wrap {
  width: min(1120px, 100% - 48px);
  margin-inline: auto;
}

/* ---------- nav ---------- */
.nav {
  position: sticky;
  top: 0;
  z-index: 10;
  background: rgba(246, 247, 251, 0.86);
  backdrop-filter: saturate(1.4) blur(10px);
  border-bottom: 1px solid var(--line);
}
.nav .wrap {
  display: flex;
  align-items: center;
  gap: 32px;
  height: 64px;
}
.brand {
  display: flex;
  align-items: center;
  gap: 10px;
  text-decoration: none;
  font: 700 19px/1 var(--latin);
  letter-spacing: -0.01em;
}
.brand-logo {
  display: block;
  width: 28px;
  height: 28px;
  border-radius: 7px;
  object-fit: contain;
}
.brand-mark {
  width: 26px;
  height: 26px;
  border-radius: 7px;
  background: var(--seat);
  display: grid;
  place-items: center;
}
.brand-mark::after {
  content: '';
  width: 10px;
  height: 10px;
  border: 2.5px solid #fff;
  border-radius: 50%;
  border-right-color: transparent;
  transform: rotate(-45deg);
}
.nav-links {
  display: flex;
  gap: 26px;
  margin-right: auto;
}
.nav-links a {
  text-decoration: none;
  font-size: 15px;
  color: var(--ink-soft);
}
.nav-links a:hover {
  color: var(--ink);
}
.nav-auth {
  display: flex;
  gap: 10px;
  align-items: center;
}

/* prefixed to avoid colliding with the app's global .btn* component classes */
.gr-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  height: 44px;
  padding: 0 20px;
  border-radius: 10px;
  font: 600 15px/1 var(--sans);
  text-decoration: none;
  white-space: nowrap;
  border: 1px solid transparent;
  transition:
    background 0.15s,
    border-color 0.15s,
    transform 0.15s;
}
.gr-btn-sm {
  height: 38px;
  padding: 0 16px;
  font-size: 14px;
}
.gr-btn-primary {
  background: var(--seat);
  color: #fff;
}
.gr-btn-primary:hover {
  background: var(--seat-deep);
}
.gr-btn-ghost {
  background: transparent;
  color: var(--ink);
  border-color: var(--line);
}
.gr-btn-ghost:hover {
  border-color: var(--ink-faint);
  background: var(--card);
}
.gr-btn-plain {
  color: var(--ink-soft);
  padding: 0 8px;
}
.gr-btn-plain:hover {
  color: var(--ink);
}

/* ---------- hero ---------- */
.hero {
  padding: 88px 0 56px;
  overflow-x: clip;
}
.hero-grid {
  display: grid;
  grid-template-columns: 1.15fr 0.85fr;
  gap: 64px;
  align-items: center;
}
.eyebrow {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: var(--verified);
  font-weight: 600;
  padding: 6px 12px;
  border-radius: 999px;
  background: var(--verified-wash);
}
.eyebrow::before {
  content: '';
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--verified);
}
h1 {
  margin: 22px 0 20px;
  font-size: clamp(36px, 5.2vw, 60px);
  line-height: 1.12;
  font-weight: 800;
  letter-spacing: -0.02em;
}
h1 .latin {
  font-family: var(--latin);
  color: var(--seat);
  letter-spacing: -0.03em;
}
.lede {
  font-size: 18px;
  color: var(--ink-soft);
  max-width: 34em;
  margin: 0 0 32px;
}
.hero-cta {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  align-items: center;
}
.hero-cta .signin {
  font-size: 14px;
  color: var(--ink-faint);
  margin-left: 6px;
}
.hero-cta .signin a {
  color: var(--seat);
  font-weight: 600;
  text-decoration: none;
}
.hero-cta .signin a:hover {
  text-decoration: underline;
}

/* ---------- signature: seat card ---------- */
.seat-card {
  position: relative;
  background: var(--card);
  border: 1px solid var(--line);
  border-radius: 20px;
  box-shadow:
    0 1px 0 #fff inset,
    0 30px 60px -28px rgba(42, 53, 168, 0.35),
    0 8px 18px -10px rgba(20, 22, 43, 0.12);
  padding: 26px 28px 22px;
  transform: rotate(-1.6deg);
  animation: settle 0.9s cubic-bezier(0.2, 0.8, 0.2, 1) both;
}
@keyframes settle {
  from {
    transform: translateY(18px) rotate(-4deg);
  }
  to {
    opacity: 1;
    transform: rotate(-1.6deg);
  }
}
.seat-card::before,
.seat-card::after {
  /* ticket notches */
  content: '';
  position: absolute;
  top: 118px;
  width: 22px;
  height: 22px;
  border-radius: 50%;
  background: var(--paper);
  border: 1px solid var(--line);
}
.seat-card::before {
  left: -12px;
  clip-path: inset(0 0 0 50%);
}
.seat-card::after {
  right: -12px;
  clip-path: inset(0 50% 0 0);
}
.seat-head {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
}
.seat-issuer {
  font: 600 12px/1.4 var(--mono);
  letter-spacing: 0.12em;
  color: var(--ink-faint);
  text-transform: uppercase;
}
.seat-status {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  font-weight: 600;
  color: var(--verified);
}
.seat-status::before {
  content: '';
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: var(--verified);
  box-shadow: 0 0 0 4px var(--verified-wash);
}
.seat-plan {
  margin: 10px 0 0;
  font: 700 26px/1.2 var(--latin);
  letter-spacing: -0.02em;
}
.seat-plan small {
  display: block;
  font: 500 13px/1.6 var(--sans);
  color: var(--ink-soft);
  letter-spacing: 0;
}
.seat-perf {
  margin: 22px -28px 18px;
  border-top: 2px dashed var(--line);
}
.seat-rows {
  margin: 0;
  display: grid;
  gap: 12px;
}
.seat-rows div {
  display: grid;
  grid-template-columns: 64px 1fr;
  gap: 12px;
  align-items: baseline;
}
.seat-rows dt {
  font: 600 12px/1.6 var(--mono);
  color: var(--ink-faint);
  letter-spacing: 0.08em;
}
.seat-rows dd {
  margin: 0;
  font-size: 15px;
  font-weight: 600;
}
.seat-foot {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  margin-top: 22px;
  padding-top: 16px;
  border-top: 1px solid var(--line);
}
.seat-no {
  font: 500 12px/1.5 var(--mono);
  color: var(--ink-faint);
}
.seat-no b {
  display: block;
  font-size: 15px;
  color: var(--ink);
  letter-spacing: 0.06em;
}
.barcode {
  width: 120px;
  height: 34px;
  background: repeating-linear-gradient(
    90deg,
    var(--ink) 0 2px,
    transparent 2px 4px,
    var(--ink) 4px 5px,
    transparent 5px 8px,
    var(--ink) 8px 11px,
    transparent 11px 12px,
    var(--ink) 12px 13px,
    transparent 13px 16px
  );
  opacity: 0.85;
}
.seat-caption {
  margin: 14px 0 1em;
  text-align: center;
  font-size: 12px;
  color: var(--ink-faint);
  transform: rotate(-1.6deg);
}

/* ---------- notice ---------- */
.notice {
  display: flex;
  gap: 14px;
  align-items: flex-start;
  padding: 16px 20px;
  margin: 56px 0 1em; /* browser default p bottom margin, as in the prototype */
  background: var(--caution-wash);
  border-radius: var(--radius);
  color: #5e4416;
  font-size: 15px;
}
.notice b {
  color: var(--caution);
  white-space: nowrap;
}

/* ---------- sections ---------- */
section.block {
  padding: 88px 0;
  border-top: 1px solid var(--line);
}
.kicker {
  font-size: 14px;
  font-weight: 600;
  color: var(--seat);
  margin: 0 0 10px;
}
h2 {
  margin: 0 0 14px;
  font-size: clamp(28px, 3.4vw, 38px);
  line-height: 1.25;
  font-weight: 800;
  letter-spacing: -0.015em;
}
.section-lede {
  margin: 0 0 40px;
  color: var(--ink-soft);
  max-width: 40em;
}

/* risk */
.policy {
  display: grid;
  grid-template-columns: auto 1fr;
  gap: 18px;
  align-items: start;
  padding: 22px 24px;
  margin-bottom: 28px;
  background: var(--card);
  border: 1px solid var(--line);
  border-left: 4px solid var(--danger);
  border-radius: var(--radius);
}
.policy h3 {
  margin: 0 0 4px;
  font-size: 17px;
  font-weight: 700;
}
.policy p {
  margin: 0;
  color: var(--ink-soft);
  font-size: 15px;
}
.policy-tag {
  font: 600 12px/1 var(--mono);
  color: var(--danger);
  padding: 6px 8px;
  border: 1px solid #ebc7c5;
  border-radius: 6px;
  margin-top: 2px;
}

.compare {
  width: 100%;
  border-collapse: separate;
  border-spacing: 0;
  background: var(--card);
  border: 1px solid var(--line);
  border-radius: var(--radius);
  overflow: hidden;
  font-size: 15px;
}
.compare th,
.compare td {
  padding: 18px 22px;
  text-align: left;
  vertical-align: top;
  border-bottom: 1px solid var(--line);
}
.compare tr:last-child td,
.compare tr:last-child th {
  border-bottom: 0;
}
.compare thead th {
  font-size: 13px;
  font-weight: 600;
  color: var(--ink-faint);
  background: #fafbfd;
}
.compare tbody th {
  font-weight: 700;
  width: 22%;
}
.compare td {
  color: var(--ink-soft);
}
.compare .ours {
  background: var(--verified-wash);
  color: #0b5e45;
  font-weight: 600;
}
.compare thead .ours {
  color: var(--verified);
  background: #ddf0e7;
}
.compare td.ours::before {
  content: '✓ ';
  color: var(--verified);
}

/* pricing */
.plans {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 24px;
  max-width: 880px;
}
.plan {
  position: relative;
  display: flex;
  flex-direction: column;
  background: var(--card);
  border: 1px solid var(--line);
  border-radius: 18px;
  padding: 30px;
}
.plan.featured {
  border: 2px solid var(--seat);
  box-shadow: 0 24px 50px -30px rgba(42, 53, 168, 0.45);
}
.plan-badge {
  position: absolute;
  top: -13px;
  left: 28px;
  background: var(--seat);
  color: #fff;
  font-size: 12px;
  font-weight: 600;
  padding: 4px 12px;
  border-radius: 999px;
}
.plan h3 {
  margin: 0;
  font: 700 22px/1.3 var(--latin);
}
.plan .ratio {
  margin: 4px 0 20px;
  color: var(--verified);
  font-size: 14px;
  font-weight: 600;
}
.price {
  display: flex;
  align-items: baseline;
  gap: 4px;
}
.price .cur {
  font: 600 20px/1 var(--latin);
  color: var(--ink-soft);
}
.price .num {
  font: 700 56px/1 var(--latin);
  letter-spacing: -0.03em;
}
.price .per {
  color: var(--ink-faint);
  font-size: 15px;
}
.was {
  margin: 6px 0 22px;
  color: var(--ink-faint);
  font-size: 14px;
  text-decoration: line-through;
}
.plan ul {
  margin: 0 0 26px;
  padding: 20px 0 0;
  border-top: 1px solid var(--line);
  list-style: none;
  flex: 1;
}
.plan li {
  position: relative;
  padding-left: 18px;
  color: var(--ink-soft);
  font-size: 15px;
}
.plan li::before {
  content: '';
  position: absolute;
  left: 0;
  top: 0.7em;
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: var(--verified);
}
.plan .gr-btn {
  width: 100%;
}
.plans-note {
  margin: 24px 0 0;
  font-size: 14px;
  color: var(--ink-faint);
}
.plans-intro {
  margin: 0 0 40px;
  color: var(--ink-soft);
  max-width: 44em;
}
.plans-intro b {
  color: var(--ink);
  font-weight: 700;
}

/* steps */
.steps {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 0;
  counter-reset: step;
  list-style: none;
  padding: 0;
  margin: 0;
}
.steps li {
  position: relative;
  padding: 0 24px 0 0;
  counter-increment: step;
}
.steps li::before {
  content: counter(step);
  display: grid;
  place-items: center;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: var(--card);
  border: 2px solid var(--seat);
  color: var(--seat);
  font: 700 15px/1 var(--latin);
  position: relative;
  z-index: 1;
  margin-bottom: 18px;
}
.steps li:not(:last-child)::after {
  content: '';
  position: absolute;
  top: 17px;
  left: 36px;
  right: 0;
  border-top: 2px dashed #c9cdf3;
}
.steps h3 {
  margin: 0 0 8px;
  font-size: 17px;
  font-weight: 700;
}
.steps p {
  margin: 0;
  color: var(--ink-soft);
  font-size: 15px;
}
.tools {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 10px;
}
.tools span {
  font: 500 12px/1 var(--mono);
  padding: 6px 8px;
  border-radius: 6px;
  background: var(--seat-wash);
  color: var(--seat-deep);
}

/* faq */
.faq {
  display: grid;
  gap: 12px;
  max-width: 880px;
}
.faq details {
  background: var(--card);
  border: 1px solid var(--line);
  border-radius: var(--radius);
}
.faq summary {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
  padding: 20px 24px;
  cursor: pointer;
  list-style: none;
  font-weight: 600;
  font-size: 16px;
}
.faq summary::-webkit-details-marker {
  display: none;
}
.faq summary::after {
  content: '+';
  font: 400 22px/1 var(--latin);
  color: var(--seat);
  transition: transform 0.2s;
}
.faq details[open] summary::after {
  transform: rotate(45deg);
}
.faq details p {
  margin: 0;
  padding: 0 24px 22px;
  color: var(--ink-soft);
  font-size: 15px;
}

/* reading */
.reading {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
}
.reading a {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 22px;
  background: var(--card);
  border: 1px solid var(--line);
  border-radius: var(--radius);
  text-decoration: none;
  transition:
    border-color 0.15s,
    transform 0.15s;
}
.reading a:hover {
  border-color: var(--seat);
  transform: translateY(-2px);
}
.reading .tag {
  align-self: flex-start;
  font-size: 12px;
  font-weight: 600;
  color: var(--verified);
  background: var(--verified-wash);
  padding: 3px 8px;
  border-radius: 6px;
}
.reading b {
  font-size: 16px;
  font-weight: 700;
}
.reading span.desc {
  color: var(--ink-soft);
  font-size: 14px;
}
.reading .todo {
  font: 500 12px/1.4 var(--mono);
  color: var(--caution);
}

/* closing CTA */
.closing {
  padding: 80px 0;
}
.closing .box {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 32px;
  flex-wrap: wrap;
  background: var(--ink);
  color: #fff;
  border-radius: 22px;
  padding: 44px 48px;
}
.closing h2 {
  color: #fff;
  margin: 0 0 6px;
  font-size: clamp(24px, 2.8vw, 32px);
}
.closing p {
  margin: 0;
  color: #b7bad3;
}
.closing .gr-btn-ghost {
  color: #fff;
  border-color: #3a3e5c;
}
.closing .gr-btn-ghost:hover {
  background: #22253f;
  border-color: #5a5e80;
}
.closing .gr-btn-primary {
  background: #fff;
  color: var(--ink);
}
.closing .gr-btn-primary:hover {
  background: var(--seat-wash);
}

footer {
  border-top: 1px solid var(--line);
  padding: 28px 0 40px;
  font-size: 14px;
  color: var(--ink-faint);
}
footer .wrap {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
}
footer a {
  text-decoration: none;
  margin-left: 18px;
}
footer a:hover {
  color: var(--ink);
}

/* ---------- responsive ---------- */
@media (max-width: 960px) {
  .hero-grid {
    grid-template-columns: 1fr;
    gap: 48px;
  }
  .seat-wrap {
    max-width: 420px;
  }
  .steps {
    grid-template-columns: 1fr 1fr;
    row-gap: 36px;
  }
  .steps li:nth-child(2)::after {
    display: none;
  }
  .reading {
    grid-template-columns: 1fr 1fr;
  }
}
@media (max-width: 720px) {
  .nav-links {
    display: none;
  }
  .nav .wrap {
    gap: 12px;
    justify-content: space-between;
  }
  .hero {
    padding-top: 56px;
  }
  .plans {
    grid-template-columns: 1fr;
    gap: 32px;
  }
  .steps {
    grid-template-columns: 1fr;
  }
  .steps li::after {
    display: none;
  }
  .reading {
    grid-template-columns: 1fr;
  }
  .closing .box {
    padding: 32px 26px;
  }
  section.block {
    padding: 64px 0;
  }
  /* comparison table → stacked cards */
  .compare,
  .compare tbody,
  .compare tr,
  .compare th,
  .compare td {
    display: block;
    width: 100%;
  }
  .compare {
    border: 0;
    background: transparent;
  }
  .compare thead {
    display: none;
  }
  .compare tr {
    background: var(--card);
    border: 1px solid var(--line);
    border-radius: var(--radius);
    margin-bottom: 12px;
    overflow: hidden;
  }
  .compare tbody th {
    width: auto;
    padding-bottom: 6px;
    border: 0;
  }
  .compare td {
    border: 0;
    padding: 4px 22px;
  }
  .compare td::before {
    content: attr(data-label) '：';
    color: var(--ink-faint);
    font-weight: 500;
  }
  .compare td.ours {
    padding: 10px 22px 14px;
    margin-top: 6px;
  }
  .compare td.ours::before {
    content: 'Goreach：✓ ';
  }
}
@media (prefers-reduced-motion: reduce) {
  .goreach,
  .goreach *,
  .goreach *::before,
  .goreach *::after {
    animation: none !important;
    transition: none !important;
  }
}
</style>
