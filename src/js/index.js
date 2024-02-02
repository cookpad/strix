// import '../css/reset.scss'
// import '../css/main.scss'
import "../css/coreui.scss";

import _ from "lodash";
import "babel-polyfill";
import * as Vue from 'vue';
import { createRouter } from 'vue-router'
import CoreuiVue from "@coreui/vue";

import Header from "./header.vue";
import Search from "./search.vue";
import Query from "./query.vue";

Vue.component("strix-header", Header);
Vue.component("strix-query", Query);
Vue.component("strix-search", Search);
Vue.use(VueRouter);
Vue.use(CoreuiVue);

import { CChartBar } from "@coreui/vue-chartjs";
Vue.component("CChartBar", CChartBar);

const router = new createRouter({
  routes: [
    {
      path: "/search/:search_id",
      component: {
        template: `<div>
        <CWrapper>
        <strix-header></strix-header>
        <div class="c-body">
        <main class="c-main">
        <strix-query></strix-query>
        <strix-search></strix-search>
        </main>
        </div>
        </CWrapper>
        </div>`,
      },
    },
    {
      path: "/",
      component: {
        template: `<div>
        <CWrapper>
        <strix-header></strix-header>
        <div class="c-body">
        <main class="c-main">
        <strix-query></strix-query>
        </main>
        </div>
        </CWrapper>
        </div>`,
      },
    },
  ],
});

const app = new Vue({
  router,
}).$mount("#app");
