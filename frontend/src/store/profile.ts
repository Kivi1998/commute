import { defineStore } from 'pinia'

export interface UserProfile {
  current_city: string
  target_position: string
  experience_years?: number
}

export const useProfileStore = defineStore('profile', {
  state: () => ({
    profile: null as UserProfile | null,
  }),
  actions: {
    setProfile(p: UserProfile) {
      this.profile = p
    },
  },
})
