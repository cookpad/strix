<template>
  <!-- Log view --->
  <div>
    <div class="row" v-if="errorMessage !== null">
      <div class="columns">
        <div class="msgbox alert">[Error] {{ errorMessage }}</div>
      </div>
      <div class="columns">
        <button class="alert-dark thin2" v-on:click="clearError()">Dismiss</button>
      </div>
    </div>

    <div class="row">
      <div class="columns metadata-view">
        <div class="metadata">
          <h3 class="metadata">Results</h3>
          <div v-if="metadata !== null">
            <div class="content">Elapsed time: {{ metadata.elapsed_seconds }} seconds</div>
            <div class="content">Total: {{ metadata.total }} logs</div>
            <div class="content">SubTotal: {{ metadata.sub_total }} logs</div>
            <div class="content">Tags: {{ metadata.tags }}</div>
          </div>
        </div>
      </div>

      <!-- TODO: add "columns" class.
    Originally the <div> also should have "column", but layout will be broken if value is too long.
      I'm not good CSS writer for now-->
      <div class="log-view">
        <div class="subrow">
          <input
            type="text"
            class="filter-query"
            autofocus
            autocomplete="off"
            placeholder="filter (write jq grammer)"
            v-model="query"
            @keyup.enter="renewJqQuery"
          />
        </div>

        <div class="subrow msgbox sysmsg" v-if="systemMessage !== null">{{ systemMessage }}</div>

        <!-- Pagenation (header) -->
        <div class="subrow" v-if="pages.length > 0">
          <ul class="pagination">
            <li v-for="p in pages" v-bind:class="{current: p.current}">
              <a
                class="offset"
                href="javascript: void(0)"
                v-on:click="changeSearchResultOffset(p.offset, p.current)"
              >{{ p.label }}</a>
            </li>
          </ul>
        </div>

        <!-- Log view -->
        <div class="subrow" v-if="logs.length > 0">
          <table class="log-view">
            <thead>
              <tr>
                <td>Meta</td>
                <td>Log</td>
              </tr>
            </thead>
            <tbody>
              <tr v-for="log in logs">
                <td class="log-meta-data">
                  <div class="content">
                    <strong>{{ log.datetime }}</strong>
                    <span class="label" v-bind:style="log.labelStyle">{{ log.tag }}</span>
                  </div>
                </td>
                <td>
                  <table class="log-data-view">
                    <tbody>
                      <tr v-for="d in log.data">
                        <td class="log-field">{{ d.k }}</td>
                        <td class="log-value">
                          <div class="log-value" v-html="d.v"></div>
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Pagenation (footer) -->
        <div class="subrow" v-if="pages.length > 0">
          <ul class="pagination">
            <li v-for="p in pages" v-bind:class="{current: p.current}">
              <a
                class="offset"
                href="javascript: void(0)"
                v-on:click="changeSearchResultOffset(p.offset)"
              >{{ p.label }}</a>
            </li>
          </ul>
        </div>
      </div>
    </div>
  </div>
</template>
<script>
import axios from "axios";
import strftime from "strftime";
import querystring from "querystring";
import SHA1 from "crypto-js/sha1";
import escapeHTML from "escape-html";
import { prototype } from "events";

var httpClient = axios.create({
  headers: { "x-api-key": localStorage.getItem("apiKey") }
});

const nowDatetime = new Date();
const utcDatetime = new Date(
  nowDatetime.getUTCFullYear(),
  nowDatetime.getUTCMonth(),
  nowDatetime.getUTCDate(),
  nowDatetime.getUTCHours(),
  nowDatetime.getUTCMinutes(),
  nowDatetime.getUTCSeconds()
);

const appData = {
  // In this component, "query" means query of jq for filtering.
  query: null,
  searchStatus: null,
  apiKey: localStorage.getItem("apiKey"),
  logs: [],
  pages: [],
  metadata: null,
  errorMessage: null,
  systemMessage: null,

  // lastQueryString will be set this.$route.query
  lastQueryString: null,
  search_id: null
};

export default {
  data() {
    return appData;
  },
  methods: {
    changeSearchResultOffset: changeSearchResultOffset,
    showSearch: showSearch,
    renewJqQuery: renewJqQuery,
    clearError: clearError
  },
  mounted() {
    this.showSearch();
  },
  watch: {
    $route(to, from) {
      console.log(from, " => ", to);
      if (to.matched[0].path === "/search/:search_id") {
        this.showSearch();
      }
    }
  }
};

// Ref: https://katashin.info/2018/12/18/247
function toTextColor(color) {
  var r = parseInt(color.substr(1, 2), 16);
  var g = parseInt(color.substr(3, 2), 16);
  var b = parseInt(color.substr(5, 2), 16);

  const toRgbItem = item => {
    const i = item / 255;
    return i <= 0.03928 ? i / 12.92 : Math.pow((i + 0.055) / 1.055, 2.4);
  };
  const R = toRgbItem(r);
  const G = toRgbItem(g);
  const B = toRgbItem(b);
  const Lbg = 0.2126 * R + 0.7152 * G + 0.0722 * B;

  const Lw = 1;
  const Lb = 0;
  const Cw = (Lw + 0.05) / (Lbg + 0.05);
  const Cb = (Lbg + 0.05) / (Lb + 0.05);

  return Cw < Cb ? "#000" : "#fff";
}

function renderLogDataValue(raw, queryTerms) {
  let values = [escapeHTML(raw)];
  const terms = queryTerms.map(x => x.term);

  terms.forEach(t => {
    values = values
      .map(x => x.split(t))
      .reduce((p, c) => p.concat(c), [])
      .map(v => [v, t])
      .reduce((p, c) => p.concat(c), [])
      .slice(0, -1);
  });

  const msg = values
    .map(v => {
      if (terms.indexOf(v) >= 0) {
        return '<span class="log-highlight">' + v + "</span>";
      } else {
        return v;
      }
    })
    .join("");

  return msg;
}

function buildPagenationIndices(metadata) {
  if (metadata.sub_total === 0) {
    return []; // No page index
  }

  const step = metadata.limit;
  const bakPages = [];
  const fwdPages = [];
  let tgtPages = bakPages;
  let lastPage;
  let currentPage;

  for (var i = 0; i * step < metadata.sub_total; i++) {
    const current = metadata.offset === i * step;
    const p = {
      label: i + 1,
      offset: i * step,
      current: current,
      is_link: true
    };

    if (current) {
      tgtPages = fwdPages;
      currentPage = p;
    } else {
      tgtPages.push(p);
    }

    lastPage = p;
  }

  const pages = [].concat(
    [
      {
        label: "<<",
        offset: 0,
        current: false,
        is_link: true
      }
    ],
    bakPages.length > 5 ? [{ label: "..." }] : [],
    bakPages.slice(-5),
    [currentPage],
    fwdPages.slice(0, 5),
    fwdPages.length > 5 ? [{ label: "..." }] : [],
    [
      {
        label: ">>",
        offset: lastPage.offset,
        current: false,
        is_link: true
      }
    ]
  );

  return pages;
}

function renderResult(data) {
  appData.searchStatus = null;
  appData.metadata = data.metadata;
  appData.pages = buildPagenationIndices(data.metadata);

  if (data.logs === null || data.logs.length === 0) {
    appData.systemMessage = "No log found";
    appData.logs = [];
    return;
  }

  appData.logs = data.logs.map(x => {
    const bgColor =
      "#" +
      SHA1(x.tag)
        .toString()
        .substring(0, 6);

    return {
      tag: x.tag,
      datetime: strftime("%F %T%z", new Date(x.timestamp * 1000)),
      data: Object.keys(x.log).map(k => {
        const v =
          typeof x.log[k] === "object"
            ? JSON.stringify(x.log[k], null, 4)
            : x.log[k];

        return {
          k: k,
          // v: renderLogDataValue(v)
          v: v
        };
      }),
      labelStyle: {
        "background-color": bgColor,
        color: toTextColor(bgColor)
      }
    };
  });
}

function showError(errMsg) {
  appData.errorMessage = errMsg;
}

function clearError() {
  appData.errorMessage = null;
}

function renewSearchResult(router, newQuery) {
  const qs = Object.assign(appData.lastQueryString, newQuery);

  const url =
    `/search/${appData.search_id}` +
    (Object.keys(qs).length > 0 ? "?" + querystring.stringify(qs) : "");
  console.log("renew => ", url);
  router.push(url);
}

function renewJqQuery() {
  renewSearchResult(this.$router, { query: appData.query });
}

function changeSearchResultOffset(offset, current) {
  if (current || offset === undefined) {
    return;
  }

  renewSearchResult(this.$router, { offset: offset });
}

function getSearchResult(search_id, qs) {
  clearError();
  appData.systemMessage = null;
  const now = new Date();

  const url =
    `/api/v1/search/${search_id}/logs` +
    (Object.keys(qs).length > 0 ? "?" + querystring.stringify(qs) : "");

  httpClient
    .get(url)
    .then(function(response) {
      console.log(response);

      switch (response.data.metadata.status) {
        case "SUCCEEDED":
          renderResult(response.data);
          break;

        case "RUNNING":
          const now = new Date();
          appData.searchStatus =
            "Elapsed time: " +
            Math.floor(response.data.metadata.elapsed_seconds * 100) / 100 +
            " seconds";

          setTimeout(function() {
            getSearchResult(search_id, qs);
          }, 1000);
          break;

        default:
          showError("Fail request: " + response.data.metadata.status);
      }
    })
    .catch(err => {
      console.log("Error: ", err, err.request, err.response);
      if (err.response && err.response.data.message) {
        showError(err.response.data.message);
      }
    });
}

function showSearch() {
  appData.search_id = this.$route.params.search_id;
  appData.query = this.$route.query.query;
  appData.lastQueryString = Object.assign({}, this.$route.query);
  getSearchResult(this.$route.params.search_id, this.$route.query);
}
</script>
<style>
</style>
