<template>
  <div class="card">
    <el-form label-width="120px" @submit.prevent>
      <el-form-item label="题目 ID">
        <el-input-number v-model="form.id" :min="1" :disabled="disableID" />
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
        <el-input v-model="tagsText" placeholder="用英文逗号分隔，例如：数组,哈希表" />
      </el-form-item>

      <el-form-item label="题目来源">
        <el-input v-model="form.source" />
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
        <el-input v-model="form.content" type="textarea" :rows="12" />
      </el-form-item>

      <el-form-item label="测试用例">
        <div class="cases">
          <div v-for="(item, index) in form.testCases" :key="index" class="case-card">
            <div class="case-head">
              <strong>Case {{ index + 1 }}</strong>
              <el-button text type="danger" @click="removeCase(index)">删除</el-button>
            </div>
            <el-input v-model="item.input" type="textarea" :rows="4" placeholder="输入" />
            <el-input v-model="item.expected" type="textarea" :rows="3" placeholder="期望输出" />
          </div>
          <el-button @click="addCase">新增测试用例</el-button>
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
import { reactive, ref, watch } from 'vue'

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
  }
})

defineEmits(['submit'])

const tagsText = ref('')
const form = reactive(createDefaultForm())

watch(
  () => props.initialValue,
  value => {
    applyForm(value)
  },
  { immediate: true, deep: true }
)

watch(tagsText, value => {
  form.tags = value.split(',').map(item => item.trim()).filter(Boolean)
})

function createDefaultForm() {
  return {
    id: 2001,
    title: '',
    difficulty: '简单',
    difficultyScore: 800,
    source: 'Admin',
    timeLimit: 1000,
    memoryLimit: 256,
    outputLimitKb: 1024,
    content: '',
    tags: ['数组', '哈希表'],
    testCases: [{ input: '', expected: '' }]
  }
}

function applyForm(value = {}) {
  const merged = {
    ...createDefaultForm(),
    ...value,
    tags: Array.isArray(value.tags) && value.tags.length > 0 ? value.tags : ['数组', '哈希表'],
    testCases: Array.isArray(value.testCases) && value.testCases.length > 0
      ? value.testCases.map(item => ({ input: item.input || '', expected: item.expected || '' }))
      : [{ input: '', expected: '' }]
  }

  Object.assign(form, merged)
  tagsText.value = form.tags.join(',')
}

function addCase() {
  form.testCases.push({ input: '', expected: '' })
}

function removeCase(index) {
  if (form.testCases.length === 1) return
  form.testCases.splice(index, 1)
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
  background: #fafbfc;
}
.case-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
