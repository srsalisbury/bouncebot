import { createRouter, createWebHistory } from 'vue-router'
import HomeView from './views/HomeView.vue'
import RoomView from './views/RoomView.vue'
import HelpView from './views/HelpView.vue'
import DailyChallengeView from './views/DailyChallengeView.vue'
import DailyGameView from './views/DailyGameView.vue'
import ShareLandingView from './views/ShareLandingView.vue'
import { config } from './config'
import { useFeatureStore } from './stores/featureStore'

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
    {
      path: '/daily',
      name: 'daily',
      component: DailyChallengeView,
    },
    {
      path: '/daily/:difficulty',
      name: 'daily-game',
      component: DailyGameView,
      props: true,
    },
    {
      path: '/share/:code',
      name: 'share',
      component: ShareLandingView,
      props: true,
    },
  ],
})

let featuresLoaded = false

router.beforeEach(async (to) => {
  if (!featuresLoaded) {
    const featureStore = useFeatureStore()
    await featureStore.fetchFeatures()
    featuresLoaded = true
  }

  if (to.name === 'daily' || to.name === 'daily-game') {
    const featureStore = useFeatureStore()
    if (!featureStore.dailyChallengeEnabled) {
      return { name: 'home' }
    }
  }
})

export default router
