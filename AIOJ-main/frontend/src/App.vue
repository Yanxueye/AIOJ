<template>
  <div id="terminal-oj">
    <NavBar v-if="showNav" />
    <main :class="{ 'with-nav': showNav }">
      <router-view />
    </main>
  </div>
</template>

<script setup>
import { computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useUserStore } from '@/stores/user'
import NavBar from '@/components/NavBar.vue'

const route = useRoute()
const userStore = useUserStore()
const showNav = computed(() => !['login', 'register'].includes(route.name))

onMounted(async () => {
  if (userStore.isLoggedIn) {
    try {
      await userStore.fetchProfile()
    } catch {}
  }
})
</script>

<style scoped>
#terminal-oj {
  min-height: 100vh;
  background: var(--bg-primary);
}
main {
  min-height: 100vh;
}
main.with-nav {
  padding-top: 60px;
}
</style>
