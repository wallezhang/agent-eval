// https://vitepress.dev/guide/custom-theme
import DefaultTheme from 'vitepress/theme'
import HomeHero from './components/HomeHero.vue'
import HomeFeatures from './components/HomeFeatures.vue'
import HomeCodeDemo from './components/HomeCodeDemo.vue'
import HomeStats from './components/HomeStats.vue'
import HomeArchitecture from './components/HomeArchitecture.vue'
import HomeCTA from './components/HomeCTA.vue'
import './custom.css'

export default {
  extends: DefaultTheme,
  enhanceApp({ app }) {
    app.component('HomeHero', HomeHero)
    app.component('HomeFeatures', HomeFeatures)
    app.component('HomeCodeDemo', HomeCodeDemo)
    app.component('HomeStats', HomeStats)
    app.component('HomeArchitecture', HomeArchitecture)
    app.component('HomeCTA', HomeCTA)
  }
}
