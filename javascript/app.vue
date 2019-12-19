<template>
  <div style="strix-main">
    <div class="row">
      <input
        type="text"
        autofocus
        autocomplete="off"
        placeholder="show your query"
        v-model="query"
        @keyup.enter="submitQuery"
      />
    </div>
    <div class="row">
      <div class="columns">
        <select v-model="spanMode">
          <option value="relative" selected>Last</option>
          <option value="absolute">Between</option>
        </select>
      </div>

      <div class="columns" v-if="spanMode == 'relative'">
        <select v-model="timeSpan" class="timespan">
          <option value="3600" selected>1 hour</option>
          <option value="7200">2 hours</option>
          <option value="14400">4 hours</option>
          <option value="28800">8 hours</option>
          <option value="86400">1 day</option>
          <option value="172800">2 day</option>
          <option value="345600">4 day</option>
          <option value="604800">1 week</option>
          <option value="1209600">2 week</option>
          <option value="2419200">4 week</option>
        </select>
      </div>

      <div class="columns" v-if="spanMode === 'absolute'">
        <input type="datetime-local" v-model="timeBegin" />
        To
        <input type="datetime-local" v-model="timeEnd" />
      </div>
      <div class="columns">
        <button class="send_query thin2" v-on:click="submitQuery">Query</button>
      </div>
      <div class="columns">
        <button
          class="secondary thin2"
          v-on:click="showApiKey"
          v-if="!showApiKeyForm"
        >Change API Key</button>
      </div>
    </div>

    <div class="row" v-if="showApiKeyForm">
      <div class="columns" style="text-align: right;">
        <h4>API Key</h4>
      </div>
      <div class="columns">
        <input type="text" v-model="apiKey" />
      </div>
      <div class="columns">
        <button class="highlight thin" v-on:click="saveApiKey">Save</button>
      </div>
    </div>

    <div class="row" v-if="searchStatus !== null">
      <div class="columns query-status">
        <div class="alert-box radius">{{ searchStatus }}</div>
      </div>
    </div>

    <div class="row" v-if="errorMessage !== null">
      <div class="columns">
        <div class="alert-box alert">[Error] {{ errorMessage }}</div>
      </div>
      <div class="columns">
        <button class="alert-dark" v-on:click="clearError()">Dismiss</button>
      </div>
    </div>

    <!-- Log view --->
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
              <a href="#" v-on:click="changeSearchResultOffset(p.offset)">{{ p.index + 1 }}</a>
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
              <a href="#" v-on:click="changeSearchResultOffset(p.offset)">{{ p.index + 1 }}</a>
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

setInterval(() => {
  console.log(appData.timeBegin);
}, 1000);

export default {
  data() {
    return appData;
  },
  methods: {
    saveApiKey: saveApiKey,
    editApiKey: editApiKey,
    showApiKey: showApiKey,
    clearError: clearError,
    submitQuery: submitQuery,
    changeSearchResultOffset: changeSearchResultOffset
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

  if (data.logs !== null && data.logs.length > 0) {
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
  } else {
    appData.errorMessage = "No log found";
  }
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
  appData.logs = [];
  const qs = querystring.stringify({
    offset: offset
  });
  httpClient
    .get(`/api/v1/search/` + appData.queryID + `/logs?` + qs)
    .then(function(response) {
      console.log(response);
      switch (response.data.metadata.status) {
        case "SUCCEEDED":
          renderResult(response.data);
          break;
        default:
          showError(
            "Fail request (Status is not SUCCEEDED): " +
              response.data.metadata.status
          );
          break;
      }
    })
    .catch(showError);
}

function getSearchResult(queryID) {
  const now = new Date();

  httpClient
    .get(`/api/v1/search/` + queryID + `/logs`)
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
function submitQuery(ev) {
  clearError();

  if (appData.apiKey === "") {
    showError("API key required");
    return;
  }

  if (appData.query === "") {
    showError("No query");
    return;
  }

  console.log("submit...", ev);
  appData.logs = [];

  const span = extractSpan();

  appData.queryTerms = appData.query.split(/\s+/).map(x => {
    return { term: x };
  });

  const body = {
    query: appData.queryTerms,
    start_dt: span.start,
    end_dt: span.end
  };
  console.log("body =>", body);

  httpClient
    .post(`/api/v1/search`, body)
    .then(function(response) {
      appData.searchStatus = "Start search";
      appData.queryID = response.data.query_id;
      setTimeout(function() {
        getSearchResult(response.data.query_id);
      }, 1000);
      console.log(response);
    })
    .catch(showError);
}
</script>
<style>
</style>
