// import '../css/reset.scss'
// import '../css/main.scss'
import '../css/coreui.scss'

import _ from 'lodash';
import "babel-polyfill";
import Vue from "vue";
import VueRouter from "vue-router"
import CoreuiVue from '@coreui/vue'

import Header from "./header.vue";
import Search from "./search.vue";
import Query from "./query.vue";

Vue.component('strix-header', Header);
Vue.component('strix-query', Query);
Vue.component('strix-search', Search);
Vue.use(VueRouter);
Vue.use(CoreuiVue);

const router = new VueRouter({
  routes: [
    {
      path: '/search/:search_id',
      component: {
        template: `<div>
        <strix-header></strix-header>
        <div class="c-body">
        <main class="c-main">
        <strix-query></strix-query>
        <strix-search></strix-search>
        </main>
        </div>
        </div>`
      },
    },
    {
      path: '/',
      component: {
        template: `<div>
        <strix-header></strix-header>
        <div class="c-body">
        <main class="c-main">
        <strix-query></strix-query>
        </main>
        </div>
        </div>`,
      },
    },
  ]
})

const app = new Vue({
  router
}).$mount('#app')
