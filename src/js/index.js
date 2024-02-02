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
import { CChartBar } from "@coreui/vue-chartjs";

const router = createRouter({
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

const app = Vue.createApp({
  router
})

app.component("strix-header", Header);
app.component("strix-query", Query);
app.component("strix-search", Search);
app.component("CChartBar", CChartBar);
app.use(VueRouter)
app.use(CoreuiVue)

app.mount("#app")