<template>
  <div class="card">
    <el-form label-width="120px" @submit.prevent>
      <el-form-item v-if="!isCreate" label="题目 ID">
        <el-tag size="large">#{{ form.id }}</el-tag>
      </el-form-item>

      <el-form-item label="标题">
        <el-input v-model="form.title" />
      </el-form-item>

      <el-form-item label="难度">
        <el-select v-model="form.difficulty" style="width: 200px">
          <el-option label="简单" value="简单" />
          <el-option label="中等" value="中等" />
          <el-option label="困难" value="困难" />
        </el-select>
      </el-form-item>

      <el-form-item label="难度分">
        <el-input-number v-model="form.difficultyScore" :min="100" :step="100" />
      </el-form-item>

      <el-form-item label="标签">
        <el-select
          v-model="form.tags"
          multiple
          filterable
          placeholder="选择算法标签"
          style="width: 100%"
        >
          <el-option-group
            v-for="group in tagGroups"
            :key="group.category"
            :label="group.category"
          >
            <el-option
              v-for="t in group.tags"
              :key="t.name"
              :label="t.name"
              :value="t.name"
            />
          </el-option-group>
        </el-select>
      </el-form-item>

      <el-form-item label="题目来源">
        <el-input v-model="form.source" />
      </el-form-item>

      <el-form-item label="状态">
        <el-select v-model="form.status" style="width: 200px">
          <el-option label="草稿" value="draft" />
          <el-option label="待审核" value="review" />
          <el-option label="已发布" value="published" />
          <el-option label="已归档" value="archived" />
        </el-select>
      </el-form-item>

      <el-form-item label="审核意见">
        <el-input v-model="form.reviewComment" type="textarea" :rows="2" />
      </el-form-item>

      <el-form-item label="时间限制">
        <el-input-number v-model="form.timeLimit" :min="1" />
        <span class="unit">ms</span>
      </el-form-item>

      <el-form-item label="内存限制">
        <el-input-number v-model="form.memoryLimit" :min="1" />
        <span class="unit">MB</span>
      </el-form-item>

      <el-form-item label="输出限制">
        <el-input-number v-model="form.outputLimitKb" :min="1" />
        <span class="unit">KB</span>
      </el-form-item>

      <el-form-item label="题面">
        <el-input v-model="form.content" type="textarea" :rows="10" />
      </el-form-item>

      <el-form-item label="约束">
        <el-input v-model="form.constraints" type="textarea" :rows="4" />
      </el-form-item>

      <el-form-item label="官方题解">
        <el-input v-model="form.editorial" type="textarea" :rows="8" />
      </el-form-item>

      <el-form-item label="公开样例">
        <div class="cases">
          <div v-for="(item, index) in form.samples" :key="`sample-${index}`" class="case-card">
            <div class="case-head">
              <strong>Sample {{ index + 1 }}</strong>
              <el-button text type="danger" @click="removeSample(index)">删除</el-button>
            </div>
            <el-input v-model="item.input" type="textarea" :rows="3" placeholder="输入" />
            <el-input v-model="item.expected" type="textarea" :rows="2" placeholder="期望输出" />
            <el-input v-model="item.explanation" type="textarea" :rows="2" placeholder="样例说明" />
          </div>
          <el-button @click="addSample">新增样例</el-button>
        </div>
      </el-form-item>

      <el-form-item label="测试用例">
        <div class="cases">
          <div v-for="(item, index) in form.testCases" :key="`case-${index}`" class="case-card">
            <div class="case-head">
              <strong>Case {{ index + 1 }}</strong>
              <el-switch v-model="item.isHidden" active-text="隐藏" inactive-text="公开" />
              <el-button text type="danger" @click="removeCase(index)">删除</el-button>
            </div>
            <el-input v-model="item.input" type="textarea" :rows="4" placeholder="输入" />
            <el-input v-model="item.expected" type="textarea" :rows="3" placeholder="期望输出" />
          </div>
          <el-button @click="addCase">新增测试用例</el-button>
        </div>
      </el-form-item>

      <el-form-item label="代码模板">
        <div class="cases">
          <div v-for="(item, index) in form.templates" :key="`template-${index}`" class="case-card">
            <div class="case-head">
              <strong>模板 {{ index + 1 }}</strong>
              <el-button text type="danger" @click="removeTemplate(index)">删除</el-button>
            </div>
            <el-select v-model="item.language" style="width: 180px">
              <el-option label="C++" value="cpp" />
              <el-option label="Python3" value="python" />
              <el-option label="Go" value="go" />
            </el-select>
            <el-input v-model="item.code" type="textarea" :rows="6" placeholder="模板代码" />
          </div>
          <el-button @click="addTemplate">新增模板</el-button>
        </div>
      </el-form-item>

      <el-form-item>
        <el-button type="primary" :loading="submitting" @click="$emit('submit')">{{ submitText }}</el-button>
        <slot name="actions" />
      </el-form-item>
    </el-form>
  </div>
</template>

<script setup>
import { reactive, ref, watch, onMounted } from 'vue'
import { tagApi } from '@/api/tag'

const props = defineProps({
  initialValue: {
    type: Object,
    default: () => ({})
  },
  submitText: {
    type: String,
    default: '保存题目'
  },
  submitting: {
    type: Boolean,
    default: false
  },
  disableID: {
    type: Boolean,
    default: false
  },
  isCreate: {
    type: Boolean,
    default: false
  }
})

defineEmits(['submit'])

const tagGroups = ref([])

onMounted(async () => {
  try {
    const res = await tagApi.getList()
    tagGroups.value = res.data?.categories || []
  } catch {
    tagGroups.value = []
  }
})

const form = reactive(createDefaultForm())

watch(
  () => props.initialValue,
  value => {
    applyForm(value)
  },
  { immediate: true, deep: true }
)

function createDefaultForm() {
  return {
    id: 0,
    title: '',
    difficulty: '简单',
    difficultyScore: 800,
    source: 'Admin',
    status: 'draft',
    reviewComment: '',
    timeLimit: 1000,
    memoryLimit: 256,
    outputLimitKb: 1024,
    content: '',
    constraints: '',
    editorial: '',
    tags: ['数组', '哈希表'],
    samples: [{ input: '', expected: '', explanation: '' }],
    testCases: [{ input: '', expected: '', isHidden: false }],
    templates: defaultTemplates()
  }
}

function defaultTemplates() {
  return [
    { language: 'cpp', code: '#include <bits/stdc++.h>\nusing namespace std;\n\nint main() {\n    return 0;\n}\n' },
    { language: 'python', code: 'import sys\ninput = sys.stdin.readline\n\ndef solve():\n    pass\n\nsolve()\n' },
    { language: 'go', code: 'package main\n\nfunc main() {\n}\n' }
  ]
}

function applyForm(value = {}) {
  const merged = {
    ...createDefaultForm(),
    ...value,
    tags: Array.isArray(value.tags) && value.tags.length > 0 ? value.tags : ['数组', '哈希表'],
    samples: normalizeArray(value.samples, { input: '', expected: '', explanation: '' }),
    testCases: normalizeArray(value.testCases, { input: '', expected: '', isHidden: false }),
    templates: normalizeArray(value.templates, null).length > 0
      ? normalizeArray(value.templates, { language: 'cpp', code: '' })
      : defaultTemplates()
  }

  Object.assign(form, merged)
}

function normalizeArray(value, fallbackItem) {
  if (!Array.isArray(value) || value.length === 0) {
    return fallbackItem ? [{ ...fallbackItem }] : []
  }
  return value.map(item => ({ ...item }))
}

function addSample() {
  form.samples.push({ input: '', expected: '', explanation: '' })
}

function removeSample(index) {
  if (form.samples.length === 1) return
  form.samples.splice(index, 1)
}

function addCase() {
  form.testCases.push({ input: '', expected: '', isHidden: true })
}

function removeCase(index) {
  if (form.testCases.length === 1) return
  form.testCases.splice(index, 1)
}

function addTemplate() {
  form.templates.push({ language: 'cpp', code: '' })
}

function removeTemplate(index) {
  if (form.templates.length === 1) return
  form.templates.splice(index, 1)
}

defineExpose({
  form,
  reset: () => applyForm()
})
</script>

<style scoped>
.unit {
  margin-left: 8px;
  color: var(--text-secondary);
}
.cases {
  display: flex;
  flex-direction: column;
  gap: 12px;
  width: 100%;
}
.case-card {
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 10px;
  background: var(--bg-hover);
}
.case-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}
</style>
