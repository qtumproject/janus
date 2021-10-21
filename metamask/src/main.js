import Vue from 'vue'
import App from './App.vue'
import router from './router'
import Web3 from'web3'
import { BootstrapVue } from 'bootstrap-vue'

// Import Bootstrap an BootstrapVue CSS files (order is important)
import 'bootstrap/dist/css/bootstrap.css'
import 'bootstrap-vue/dist/bootstrap-vue.css'

Vue.config.productionTip = false
// Make BootstrapVue available throughout your project
Vue.use(BootstrapVue)
if (window.web3) {
  Vue.prototype.Web3 = new Web3(window.web3.currentProvider);
}

new Vue({
  router,
  render: function (h) { return h(App) },
}).$mount('#app')
