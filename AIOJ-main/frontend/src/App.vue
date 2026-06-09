<template>
  <div id="terminal-oj">
    <NavBar v-if="showNav" />
    <main :class="{ 'with-nav': showNav }">
      <router-view />
    </main>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import NavBar from '@/components/NavBar.vue'

const route = useRoute()
const showNav = computed(() => !['login', 'register'].includes(route.name))
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
