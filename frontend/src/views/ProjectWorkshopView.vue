<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { projectApi, courseApi } from '@/api/client'
import type { Project, ProjectFile, ProjectTemplate, Course } from '@/types'

const route = useRoute()
const router = useRouter()

const templates = ref<ProjectTemplate[]>([])
const courses = ref<Course[]>([])
const projects = ref<Project[]>([])
const loading = ref(true)

// 编辑器状态
const editing = ref(false)
const current = ref<Project | null>(null)
const files = ref<ProjectFile[]>([])
const activeFile = ref('')
const running = ref(false)
const runOutput = ref('')
const runError = ref('')
const runTime = ref(0)
const saving = ref(false)
const message = ref('')
const error = ref('')

// 新建表单
const showCreate = ref(false)
const newCourseId = ref<number | undefined>(undefined)
const newTitle = ref('')
const newDesc = ref('')
const newLang = ref('python')

onMounted(async () => {
  const projectId = Number(route.query.editor)
  try {
    const [tplRes, courseRes, listRes] = await Promise.all([
      projectApi.templates(),
      courseApi.list(),
      projectApi.list(),
    ])
    templates.value = tplRes.data.templates
    courses.value = courseRes.data.courses
    projects.value = listRes.data.projects
    if (projectId > 0) {
      await openProject(projectId)
    }
  } catch (e) {
    // ignore
  } finally {
    loading.value = false
  }
})

const activeContent = computed({
  get: () => files.value.find((f) => f.name === activeFile.value)?.content ?? '',
  set: (v: string) => {
    const f = files.value.find((f) => f.name === activeFile.value)
    if (f) f.content = v
  },
})

function fileTabs() {
  return [...files.value].sort((a, b) => a.order - b.order)
}

function setActive(name: string) {
  activeFile.value = name
}

async function openProject(id: number) {
  try {
    const res = await projectApi.get(id)
    current.value = res.data.project
    files.value = res.data.files.map((f, i) => ({ ...f, order: f.order || i }))
    activeFile.value = [...files.value].sort((a, b) => a.order - b.order)[0]?.name || ''
    runOutput.value = res.data.project.last_run_output || ''
    runError.value = res.data.project.last_run_error || ''
    editing.value = true
    error.value = ''
  } catch (e) {
    error.value = '打开项目失败'
  }
}

function closeEditor() {
  editing.value = false
  current.value = null
  files.value = []
  runOutput.value = ''
  runError.value = ''
  router.replace('/projects')
}

async function createProject() {
  if (!newTitle.value.trim()) {
    error.value = '请输入项目名称'
    return
  }
  message.value = ''
  error.value = ''
  try {
    const res = await projectApi.create({
      course_id: newCourseId.value,
      title: newTitle.value,
      description: newDesc.value,
      language: newLang.value,
    })
    projects.value.unshift(res.data)
    showCreate.value = false
    newTitle.value = ''
    newDesc.value = ''
    await openProject(res.data.id)
  } catch (e) {
    error.value = '创建项目失败'
  }
}

async function saveFiles() {
  if (!current.value) return
  saving.value = true
  message.value = ''
  try {
    const payload = files.value.map((f, i) => ({ ...f, order: i }))
    await projectApi.saveFiles(current.value.id, current.value.main_file, payload)
    message.value = '✅ 已保存'
  } catch (e) {
    error.value = '保存失败'
  } finally {
    saving.value = false
  }
}

async function runProject() {
  if (!current.value) return
  running.value = true
  runOutput.value = ''
  runError.value = ''
  message.value = ''
  try {
    const res = await projectApi.run(current.value.id)
    runOutput.value = res.data.output || ''
    runError.value = res.data.error || ''
    runTime.value = res.data.time_ms
    if (runError.value) error.value = runError.value
    else message.value = `✅ 运行成功（${res.data.time_ms}ms）`
    if (current.value) {
      current.value.run_count++
      current.value.last_run_output = runOutput.value
      current.value.last_run_error = runError.value
    }
  } catch (e) {
    error.value = '运行失败'
  } finally {
    running.value = false
  }
}

async function completeProject() {
  if (!current.value) return
  if (!confirm('确认完成该项目？完成后将解锁相关成就。')) return
  try {
    const res = await projectApi.complete(current.value.id)
    current.value = res.data.project
    message.value = '🎉 项目已完成！'
  } catch (e) {
    error.value = '操作失败'
  }
}

async function deleteProject(id: number) {
  if (!confirm('确认删除该项目？所有文件将被清除。')) return
  try {
    await projectApi.remove(id)
    projects.value = projects.value.filter((p) => p.id !== id)
    if (editing.value && current.value?.id === id) closeEditor()
  } catch (e) {
    error.value = '删除失败'
  }
}
</script>

<template>
  <div class="workshop-page">
    <div class="header">
      <router-link to="/" class="back-btn">← 返回</router-link>
      <h1>🛠️ 项目实战工坊</h1>
    </div>

    <!-- ===== 列表模式 ===== -->
    <template v-if="!editing">
      <div v-if="loading" class="loading">加载中...</div>

      <template v-else>
        <div class="toolbar">
          <button class="btn-primary" @click="showCreate = !showCreate">
            {{ showCreate ? '收起' : '＋ 新建项目' }}
          </button>
        </div>

        <div v-if="showCreate" class="create-form">
          <h2 class="section-title">新建项目</h2>
          <label class="field">
            <span class="field-label">项目名称 *</span>
            <input v-model="newTitle" class="input" placeholder="例如：猜数字小游戏" />
          </label>
          <label class="field">
            <span class="field-label">项目描述</span>
            <input v-model="newDesc" class="input" placeholder="一句话描述你的项目" />
          </label>
          <label class="field">
            <span class="field-label">语言</span>
            <select v-model="newLang" class="input">
              <option v-for="t in templates" :key="t.language" :value="t.language">{{ t.title }}</option>
            </select>
          </label>
          <label class="field">
            <span class="field-label">关联课程（可选）</span>
            <select v-model="newCourseId" class="input">
              <option :value="undefined">不关联</option>
              <option v-for="c in courses" :key="c.id" :value="c.id">{{ c.title }}</option>
            </select>
          </label>
          <p v-if="error" class="error">{{ error }}</p>
          <button class="btn-primary" @click="createProject">创建并打开</button>
        </div>

        <div class="project-list">
          <div v-if="projects.length === 0" class="empty">
            还没有项目，点击「新建项目」创建你的第一个作品吧！
          </div>
          <div v-for="p in projects" :key="p.id" class="project-card">
            <div class="project-head">
              <span class="project-icon">{{ p.status === 'completed' ? '✅' : '📝' }}</span>
              <div class="project-body">
                <div class="project-title">{{ p.title }}</div>
                <div class="project-meta">
                  {{ p.language }} · {{ p.status === 'completed' ? '已完成' : '进行中' }}
                  <span v-if="p.run_count > 0"> · 运行 {{ p.run_count }} 次</span>
                </div>
              </div>
            </div>
            <p v-if="p.description" class="project-desc">{{ p.description }}</p>
            <div class="project-actions">
              <button class="btn-secondary" @click="openProject(p.id)">打开编辑</button>
              <button class="btn-ghost" @click="deleteProject(p.id)">删除</button>
            </div>
          </div>
        </div>
      </template>
    </template>

    <!-- ===== 编辑器模式 ===== -->
    <template v-else-if="current">
      <div class="editor-head">
        <button class="back-btn" @click="closeEditor">← 列表</button>
        <h2 class="editor-title">{{ current.title }}</h2>
        <span class="editor-status" :class="{ done: current.status === 'completed' }">
          {{ current.status === 'completed' ? '已完成' : '进行中' }}
        </span>
      </div>

      <div class="editor-tabs">
        <button
          v-for="f in fileTabs()"
          :key="f.name"
          class="file-tab"
          :class="{ active: f.name === activeFile }"
          @click="setActive(f.name)"
        >{{ f.name }}</button>
        <span class="editor-lang">{{ current.language }}</span>
      </div>

      <textarea
        v-model="activeContent"
        class="code-editor"
        :placeholder="`编辑 ${activeFile} 的代码...`"
        spellcheck="false"
      ></textarea>

      <div class="editor-actions">
        <button class="btn-primary" :disabled="running || saving" @click="runProject">
          {{ running ? '运行中...' : '▶ 运行' }}
        </button>
        <button class="btn-secondary" :disabled="saving" @click="saveFiles">
          {{ saving ? '保存中...' : '💾 保存' }}
        </button>
        <button
          v-if="current.status !== 'completed'"
          class="btn-complete"
          @click="completeProject"
        >🏁 标记完成</button>
      </div>

      <p v-if="message" class="success">{{ message }}</p>
      <p v-if="error" class="error">{{ error }}</p>

      <div class="run-panel">
        <div class="run-head">
          <span>运行输出</span>
          <span v-if="runTime" class="run-time">{{ runTime }}ms</span>
        </div>
        <pre class="run-output" :class="{ err: !!runError }">{{ runOutput || runError || '点击「运行」查看输出结果' }}</pre>
      </div>
    </template>
  </div>
</template>

<style scoped>
.workshop-page { max-width: 700px; margin: 0 auto; padding: 20px; }
.header { display: flex; align-items: center; gap: 16px; margin-bottom: 24px; }
.back-btn { color: var(--text-light); text-decoration: none; font-size: 16px; cursor: pointer; border: none; background: none; font-family: inherit; }
.header h1 { font-size: 28px; font-weight: 900; }
.loading, .empty { text-align: center; padding: 40px 20px; color: var(--text-light); }

.toolbar { margin-bottom: 16px; }
.btn-primary, .btn-secondary, .btn-complete {
  padding: 10px 18px;
  border: none;
  border-radius: var(--radius-sm);
  font-size: 14px;
  font-weight: 800;
  cursor: pointer;
}
.btn-primary { background: var(--primary); color: white; box-shadow: 0 3px 0 var(--primary-dark); }
.btn-primary:disabled { opacity: 0.6; cursor: not-allowed; }
.btn-secondary { background: white; color: var(--text); border: 2px solid var(--border); box-shadow: 0 3px 0 #e5e5e5; }
.btn-complete { background: var(--secondary); color: white; box-shadow: 0 3px 0 #b45309; }
.btn-ghost { background: none; border: none; color: var(--text-light); font-size: 13px; cursor: pointer; }

.create-form {
  background: white;
  border: 2px solid var(--border);
  border-radius: var(--radius-sm);
  padding: 20px;
  margin-bottom: 16px;
}
.section-title { font-size: 16px; font-weight: 800; margin-bottom: 14px; }
.field { display: block; margin-bottom: 12px; }
.field-label { display: block; font-size: 13px; font-weight: 700; color: var(--text-light); margin-bottom: 6px; }
.input {
  width: 100%;
  padding: 10px 12px;
  border: 2px solid var(--border);
  border-radius: 10px;
  font-size: 14px;
  font-family: inherit;
  box-sizing: border-box;
}
.input:focus { outline: none; border-color: var(--primary); }

.project-list { display: flex; flex-direction: column; gap: 12px; }
.project-card {
  background: white;
  border: 2px solid var(--border);
  border-radius: var(--radius-sm);
  padding: 16px;
}
.project-head { display: flex; align-items: center; gap: 12px; }
.project-icon { font-size: 24px; }
.project-body { flex: 1; min-width: 0; }
.project-title { font-size: 16px; font-weight: 800; }
.project-meta { font-size: 12px; color: var(--text-light); }
.project-desc { font-size: 13px; color: var(--text-light); margin: 8px 0; }
.project-actions { display: flex; gap: 10px; margin-top: 10px; }
.project-actions .btn-secondary { flex: 1; text-align: center; }

/* 编辑器 */
.editor-head { display: flex; align-items: center; gap: 12px; margin-bottom: 12px; }
.editor-title { font-size: 18px; font-weight: 900; flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.editor-status {
  font-size: 12px;
  font-weight: 700;
  color: var(--warning);
  background: #fffbeb;
  padding: 3px 10px;
  border-radius: 12px;
}
.editor-status.done { color: #10b981; background: #ecfdf5; }

.editor-tabs {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 8px;
  overflow-x: auto;
}
.file-tab {
  border: 2px solid var(--border);
  background: white;
  border-radius: 8px 8px 0 0;
  padding: 8px 14px;
  font-size: 13px;
  font-weight: 700;
  cursor: pointer;
  font-family: monospace;
  border-bottom: none;
}
.file-tab.active { border-color: var(--primary); color: var(--primary); background: rgba(74,137,255,0.06); }
.editor-lang { margin-left: auto; font-size: 12px; color: var(--text-light); }

.code-editor {
  width: 100%;
  height: 260px;
  padding: 12px;
  border: 2px solid var(--border);
  border-radius: 10px;
  font-family: 'JetBrains Mono', Consolas, monospace;
  font-size: 13px;
  line-height: 1.6;
  box-sizing: border-box;
  resize: vertical;
  background: #0f172a;
  color: #e2e8f0;
}
.code-editor:focus { outline: none; border-color: var(--primary); }

.editor-actions { display: flex; gap: 10px; margin-top: 12px; flex-wrap: wrap; }

.success { color: #10b981; font-size: 13px; margin-top: 10px; font-weight: 700; }
.error { color: var(--danger); font-size: 13px; margin-top: 10px; }

.run-panel {
  margin-top: 16px;
  background: #0f172a;
  border-radius: 10px;
  overflow: hidden;
}
.run-head {
  display: flex;
  justify-content: space-between;
  padding: 8px 14px;
  background: #1e293b;
  color: #94a3b8;
  font-size: 12px;
  font-weight: 700;
}
.run-time { font-family: monospace; }
.run-output {
  margin: 0;
  padding: 14px;
  color: #e2e8f0;
  font-family: 'JetBrains Mono', Consolas, monospace;
  font-size: 13px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-all;
  min-height: 60px;
  max-height: 240px;
  overflow-y: auto;
}
.run-output.err { color: #fca5a5; }
</style>
