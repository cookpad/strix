<template>
  <!-- Log view --->
  <div>
    <div class="row" v-if="logs.length > 0">
      <div class="columns metadata-view">
        <div class="metadata">
          <h3 class="metadata">Results</h3>
          <div class="content">Elapsed time: {{ metadata.elapsed_seconds }} seconds</div>
          <div class="content">Total: {{ metadata.total }} logs</div>
        </div>
      </div>

      <!-- TODO: add "columns" class.
    Originally the <div> also should have "column", but layout will be broken if value is too long.
      I'm not good CSS writer for now-->
      <div class="log-view">
        <!-- Pagenation (header) -->
        <div v-if="pages.length > 0">
          <ul class="pagination">
            <li v-for="p in pages" v-bind:class="{current: p.current}">
              <a
                class="offset"
                href="javascript: void(0)"
                v-on:click="changeSearchResultOffset(p.offset)"
              >{{ p.index + 1 }}</a>
            </li>
          </ul>
        </div>

        <!-- Log view -->
        <div>
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
        <div v-if="pages.length > 0">
          <ul class="pagination">
            <li v-for="p in pages" v-bind:class="{current: p.current}">
              <a
                class="offset"
                href="javascript: void(0)"
                v-on:click="changeSearchResultOffset(p.offset)"
              >{{ p.index + 1 }}</a>
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
  query: "",
  queryTerms: [],
  searchStatus: null,
  queryID: null,
  apiKey: localStorage.getItem("apiKey"),
  showApiKeyForm: localStorage.getItem("apiKey") === null,
  logs: [],
  pages: [],
  metadata: {},
  timeSpan: 3600,
  timeBegin: strftime("%Y-%m-%dT%H:%M", utcDatetime),
  timeEnd: strftime("%Y-%m-%dT%H:%M", utcDatetime),
  errorMessage: null,
  spanMode: "relative"
};

export default {
  data() {
    return appData;
  },
  methods: {
    changeSearchResultOffset: changeSearchResultOffset,
    showSearch: showSearch
  },
  mounted() {
    this.showSearch();
  },
  watch: {
    $route(to, from) {
      console.log(from, "=>", to);
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

function renderLogDataValue(raw) {
  let values = [escapeHTML(raw)];
  const terms = appData.queryTerms.map(x => x.term);

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

function renderResult(data) {
  appData.searchStatus = null;
  appData.metadata = data.metadata;

  if (data.logs === null || data.logs.length === 0) {
    appData.errorMessage = "No log found";
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
          v: renderLogDataValue(v)
        };
      }),
      labelStyle: {
        "background-color": bgColor,
        color: toTextColor(bgColor)
      }
    };
  });

  appData.pages = [];
  const step = data.metadata.limit;
  for (var i = 0; i * step < data.metadata.total; i++) {
    appData.pages.push({
      index: i,
      offset: i * step,
      current: data.metadata.offset === i * step
    });
  }
  console.log("pages:", appData.pages);
}

function editApiKey(ev) {
  appData.showApiKeyForm = true;
}

function saveApiKey(ev) {
  appData.showApiKeyForm = false;
  localStorage.setItem("apiKey", appData.apiKey);
  httpClient = axios.create({
    headers: { "x-api-key": appData.apiKey }
  });
}

function showApiKey() {
  appData.showApiKeyForm = true;
}

function showError(errMsg) {
  appData.errorMessage = errMsg;
}

function clearError() {
  appData.errorMessage = null;
}

function changeSearchResultOffset(offset) {
  getSearchResult(this.$route.params.search_id, offset);
}

function getSearchResult(queryID, offset) {
  console.log("qurey =>", queryID, offset);
  appData.logs = [];
  const now = new Date();

  const qs = querystring.stringify({
    offset: offset
  });
  httpClient
    .get(`/api/v1/search/` + queryID + `/logs?` + qs)
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
            getSearchResult(queryID);
          }, 1000);
          break;

        default:
          showError("Fail request: " + response.data.metadata.status);
      }
    })
    .catch(showError);
}

function extractSpan() {
  switch (appData.spanMode) {
    case "relative": {
      const span = parseInt(appData.timeSpan);
      const now = new Date();
      const end = new Date(
        now.getUTCFullYear(),
        now.getUTCMonth(),
        now.getUTCDate(),
        now.getUTCHours(),
        now.getUTCMinutes(),
        now.getUTCSeconds()
      );

      const start = new Date(end.getTime() - span * 1000);
      const end_dt = strftime("%FT%T", end);
      const start_dt = strftime("%FT%T", start);

      return {
        start: start_dt,
        end: end_dt
      };
    }

    case "absolute": {
      const start = new Date(appData.timeBegin);
      const end = new Date(appData.timeEnd);

      return {
        start: strftime("%FT%T", start),
        end: strftime("%FT%T", end)
      };
    }
  }
}

function showSearch() {
  console.log("params =>", this.$route);
  getSearchResult(this.$route.params.search_id, 0);
}
</script>
<style>
</style>
