import '../css/reset.scss'
import '../css/main.scss'

import _ from 'lodash';
import "babel-polyfill";
import Vue from "vue";
import VueRouter from "vue-router"

import Header from "./header.vue";
import Search from "./search.vue";
import Query from "./query.vue";

Vue.component('strix-header', Header);
Vue.component('strix-query', Query);
Vue.component('strix-search', Search);
Vue.use(VueRouter);

const router = new VueRouter({
  routes: [
    {
      path: '/search/:search_id',
      component: {
        template: `<div>
        <strix-header></strix-header>
        <strix-query></strix-query>
        <strix-search></strix-search>
        </div>`
      },
    },
    {
      path: '/',
      component: {
        template: `<div>
        <strix-header></strix-header>
        <strix-query></strix-query>
        </div>`,
      },
    },
  ]
})

const app = new Vue({
  router
}).$mount('#app')
