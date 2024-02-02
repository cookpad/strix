// import '../css/reset.scss'
// import '../css/main.scss'
import "../css/coreui.scss";
import _ from "lodash";
import "babel-polyfill";
import * as Vue from 'vue';
import { createRouter, createWebHashHistory } from 'vue-router'
import { CHeaderNav, CNavLink, CNavItem, CFormInput, CFormSelect, CButton, CDropdownHeader, CDropdownItem, CDropdown, CHeader, CCol, CRow, CCardBody, CCard, CContainer   } from "@coreui/vue";
import Header from "./header.vue";
import Search from "./search.vue";
import Query from "./query.vue";
import { CChartBar } from "@coreui/vue-chartjs";

const app = Vue.createApp({});

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    {
      path: "/search/:search_id",
      component: {
        template: `<div class="wrapper d-flex flex-column min-vh-100 bg-light">
        <strix-header></strix-header>
        <div class="c-body">
        <main class="c-main">
        <strix-query></strix-query>
        <strix-search></strix-search>
        </main>
        </div>
        </div>`,
      },
    },
    {
      path: "/",
      component: {
        template: `<div class="wrapper d-flex flex-column min-vh-100 bg-light">
        <strix-header></strix-header>
        <div class="c-body">
        <main class="c-main">
        <strix-query></strix-query>
        </main>
        </div>
        </div>`,
      },
    },
  ],
});

app.component("strix-header", Header);
app.component("strix-query", Query);
app.component("strix-search", Search);
app.component("CChartBar", CChartBar);
app.component("CHeaderNav", CHeaderNav);
app.component("CNavLink", CNavLink);
app.component("CNavItem", CNavItem);
app.component("CFormInput", CFormInput);
app.component("CFormSelect", CFormSelect);
app.component("CButton", CButton);
app.component("CDropdownHeader", CDropdownHeader);
app.component("CDropdownItem", CDropdownItem);
app.component("CDropdown", CDropdown);
app.component("CHeader", CHeader);
app.component("CCol", CCol);
app.component("CRow", CRow);
app.component("CCardBody", CCardBody);
app.component("CCard", CCard);
app.component("CContainer", CContainer);


app.use(router)

app.mount("#app")