import { createRouter, createWebHistory } from 'vue-router'
import HomeView from './views/HomeView.vue'
import RoomView from './views/RoomView.vue'
import HelpView from './views/HelpView.vue'
import { config } from './config'

const router = createRouter({
  history: createWebHistory(config.basePath),
  routes: [
    {
      path: '/',
      name: 'home',
      component: HomeView,
    },
    {
      path: '/room/:roomId',
      name: 'room',
      component: RoomView,
      props: true,
    },
    {
      path: '/help',
      name: 'help',
      component: HelpView,
    },
  ],
})

export default router
