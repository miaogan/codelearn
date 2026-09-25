<script setup lang="ts">
import { useRouter } from 'vue-router'
const router = useRouter()
</script>

<template>
  <div class="container legal-page">
    <h1 class="page-title">📄 用户协议与隐私政策</h1>
    <p class="page-sub">更新日期：2026-09-25 · 适用于 CodeLearn 全部服务</p>

    <div class="card">
      <h2 class="sec-title">一、用户协议</h2>
      <p><strong>1. 账号注册</strong>：你需提供真实、准确的信息完成注册，并妥善保管账号密码；账号仅限本人使用，不得转让或出借。</p>
      <p><strong>2. 学习内容</strong>：课程、题库与练习内容仅供个人学习使用，未经许可不得复制、传播或用于商业用途。</p>
      <p><strong>3. 禁止行为</strong>：不得利用平台从事任何违法违规活动，不得提交恶意代码、干扰判题服务，不得攻击平台系统。</p>
      <p><strong>4. 能力认证</strong>：认证考试与证书用于证明学习成果，不代表任何官方职业资格。任何作弊行为将导致证书失效。</p>
      <p><strong>5. 免责声明</strong>：平台内容力求准确，但不保证完全无错；AI 生成内容仅供参考，请以官方文档与权威资料为准。</p>

      <h2 class="sec-title">二、隐私政策</h2>
      <p><strong>1. 我们收集的信息</strong>：注册信息（用户名、邮箱）、学习行为数据（课时进度、练习与考试记录、错题）、设备与日志信息（IP、浏览器类型）。</p>
      <p><strong>2. 信息的使用</strong>：用于提供个性化学习推荐、能力评估（能力图谱）、排行榜与学习日历，以及改进服务质量。</p>
      <p><strong>3. 信息的存储与保护</strong>：数据存储于加密服务器，我们采取合理的安全措施防止数据泄露、丢失或被篡改。</p>
      <p><strong>4. 信息共享</strong>：我们不会出售你的个人信息；仅在法律法规要求或为保障服务安全时向有权机关披露。</p>

      <h2 class="sec-title">三、AI 内容生成标识</h2>
      <p>平台部分习题与解答由 AI 生成（页面将标注「AI 生成」）。AI 内容可能存在错误，请结合官方文档与权威资料甄别使用；如发现错误内容，欢迎通过「问题反馈」告知我们。</p>
      <p>AI 仅用于内容辅助生成，考试组卷、判分与证书颁发均不依赖 AI，保障评价的客观性。</p>
    </div>

    <button class="btn-ghost back-btn" @click="router.back()">← 返回</button>
  </div>
</template>

<style scoped>
.legal-page { padding-top: 24px; padding-bottom: 40px; }
.page-title { font-size: 24px; margin-bottom: 4px; }
.page-sub { color: var(--text-light); font-size: 13px; margin-bottom: 20px; }
.card {
  background: white;
  border: 2px solid var(--border);
  border-radius: var(--radius);
  padding: 24px 28px;
  box-shadow: 0 4px 0 #e5e5e5;
}
.sec-title { font-size: 17px; margin: 20px 0 10px; color: var(--primary-dark); }
.sec-title:first-child { margin-top: 0; }
.card p { font-size: 14px; line-height: 1.9; color: var(--text); margin-bottom: 10px; }
.btn-ghost { background: transparent; color: var(--text-light); border: none; margin-top: 20px; width: 100%; font-weight: 800; cursor: pointer; }
</style>
