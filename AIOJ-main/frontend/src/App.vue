<template>
  <div id="terminal-oj">
    <!-- Global anime background -->
    <div class="bg-layer" />
    <div class="bg-mesh" />

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
  position: relative;
  min-height: 100vh;
}

/* ── Fixed anime character background — full screen ── */
.bg-layer {
  position: fixed;
  inset: 0;
  z-index: 0;
  pointer-events: none;
  background-image: url('/muzimi.jpg');
  background-repeat: no-repeat;
  background-position: center center;
  background-size: cover;
  opacity: 0.28;
  filter: saturate(0.85) brightness(1.02);
}

/* ── Ambient color mesh overlay ── */
.bg-mesh {
  position: fixed;
  inset: 0;
  z-index: 0;
  pointer-events: none;
  background:
    radial-gradient(ellipse at 15% 80%, rgba(82, 196, 26, 0.05) 0%, transparent 50%),
    radial-gradient(ellipse at 85% 15%, rgba(232, 168, 56, 0.04) 0%, transparent 50%),
    radial-gradient(ellipse at 50% 50%, rgba(138, 92, 196, 0.03) 0%, transparent 60%);
}

[data-theme="dark"] .bg-layer {
  opacity: 0.18;
  filter: saturate(0.6) brightness(0.75);
}

[data-theme="dark"] .bg-mesh {
  background:
    radial-gradient(ellipse at 15% 80%, rgba(82, 196, 26, 0.03) 0%, transparent 50%),
    radial-gradient(ellipse at 85% 15%, rgba(232, 168, 56, 0.02) 0%, transparent 50%);
}

main {
  position: relative;
  z-index: 1;
  min-height: 100vh;
}

main.with-nav {
  padding-top: 60px;
}

/* ── Ensure all cards/content sit above the background ── */
:deep(.card),
:deep(.el-table),
:deep(.el-dialog),
:deep(.el-dropdown-menu) {
  position: relative;
  z-index: 1;
}
</style>
