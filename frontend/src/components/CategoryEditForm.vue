<template>
  <div class="form-sections">
    <section class="form-section">
      <header class="form-section__head">
        <h3 class="form-section__title">基本信息</h3>
      </header>
      <div class="form-section__body">
        <IconPicker v-model="form.icon" title="分类图标" fallback="mdi-shape" />
        <v-text-field v-model="form.name" label="名称" variant="outlined" hide-details class="field" />
        <v-select
          v-if="!categoryId"
          v-model="form.kind"
          :items="[
            { title: '支出', value: 'expense' },
            { title: '收入', value: 'income' },
          ]"
          label="类型"
          variant="outlined"
          hide-details
          class="field"
        />
        <v-select
          v-model="form.parentId"
          :items="parentOptions"
          item-title="name"
          item-value="id"
          label="父分类（可选）"
          clearable
          variant="outlined"
          hide-details
          class="field"
          :disabled="!!categoryId && hasChildren"
        />
        <v-text-field
          v-model.number="form.sort"
          label="排序"
          type="number"
          variant="outlined"
          hide-details
          class="field field--last"
        />
        <p v-if="categoryId && hasChildren" class="form-section__tip">
          该分类下已有子分类，不可再挂到其他父分类下。
        </p>
      </div>
    </section>

    <div v-if="categoryId" class="danger-zone">
      <div class="danger-title">危险操作</div>
      <p class="danger-tip">删除分类不可恢复；若有子分类或流水将无法删除。</p>
      <v-btn color="error" variant="outlined" block @click="emit('delete')">删除此分类</v-btn>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import IconPicker from './IconPicker.vue'
import { useCategoryStore, type Category } from '../stores/category'

const props = defineProps<{
  categoryId?: number | null
  initial?: Category | null
  /** 新建时默认类型 */
  defaultKind?: 'expense' | 'income'
  /** 新建时默认父分类 */
  defaultParentId?: number | null
}>()

const emit = defineEmits<{ saved: []; delete: [] }>()

const categories = useCategoryStore()
const saving = ref(false)
const form = reactive({
  name: '',
  kind: 'expense' as 'expense' | 'income',
  icon: 'mdi-shape',
  sort: 0,
  parentId: null as number | null,
})

const hasChildren = computed(() =>
  !!props.categoryId && categories.list.some((c) => c.parentId === props.categoryId),
)

const parentOptions = computed(() => [
  { id: null as any, name: '无（一级分类）' },
  ...categories.list.filter(
    (c) =>
      c.kind === form.kind &&
      !c.parentId &&
      (!props.categoryId || c.id !== props.categoryId),
  ),
])

function fillFrom(c: Category | null | undefined) {
  if (!c) {
    form.name = ''
    form.kind = props.defaultKind || 'expense'
    form.icon = 'mdi-shape'
    form.sort = 0
    form.parentId = props.defaultParentId ?? null
    return
  }
  form.name = c.name
  form.kind = c.kind
  form.icon = c.icon || 'mdi-shape'
  form.sort = c.sort
  form.parentId = c.parentId || null
}

async function load() {
  await categories.load()
  if (props.initial) fillFrom(props.initial)
  else if (props.categoryId) {
    const c = categories.list.find((x) => x.id === props.categoryId)
    fillFrom(c || null)
  } else fillFrom(null)
}

watch(() => [props.categoryId, props.initial, props.defaultKind, props.defaultParentId], load)
onMounted(load)

async function save() {
  if (saving.value) return
  saving.value = true
  try {
    const payload: any = {
      name: form.name,
      icon: form.icon || 'mdi-shape',
      sort: form.sort,
      parentId: form.parentId || 0,
    }
    if (props.categoryId) await categories.update(props.categoryId, payload)
    else await categories.create({ ...payload, kind: form.kind, parentId: form.parentId || undefined })
    emit('saved')
  } finally {
    saving.value = false
  }
}

defineExpose({ save, saving })
</script>
