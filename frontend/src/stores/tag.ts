import { defineStore } from 'pinia'
import { ref } from 'vue'
import http from '../api/http'

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
    return data as Tag[]
  }

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

  return { list, load, create, update, remove }
})
