import { ref } from 'vue'

export const isLoggedIn = ref(false)

export function login(username: string, password: string): boolean {
  if (username.trim().length < 3 || password.length < 4) return false
  isLoggedIn.value = true
  return true
}

export function logout(): void {
  isLoggedIn.value = false
}
