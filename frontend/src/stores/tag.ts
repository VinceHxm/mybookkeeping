import { defineStore } from 'pinia'
import { ref } from 'vue'
import http from '../api/http'
import { createDictCache } from './dictCache'

export interface Tag {
  id: number
  name: string
  color: string
  sort: number
}

export const useTagStore = defineStore('tag', () => {
  const list = ref<Tag[]>([])

  async function load() {
    const { data } = await http.get('/tags')
    list.value = data
    cache.markLoaded()
    return data as Tag[]
  }
  const cache = createDictCache(load, () => list.value)

  async function create(payload: Partial<Tag>) {
    const { data } = await http.post('/tags', payload)
    await load()
    return data
  }

  async function update(id: number, payload: Partial<Tag>) {
    const { data } = await http.put(`/tags/${id}`, payload)
    await load()
    return data
  }

  async function remove(id: number) {
    await http.delete(`/tags/${id}`)
    await load()
  }

  return { list, load, ensure: cache.ensure, create, update, remove }
})
