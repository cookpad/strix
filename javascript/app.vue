<template>
  <div style="padding-top:30px;padding-bottom: 30px;">
    <div class="row">
      <div class="column large-12 query-input">
        <input
          type="text"
          autofocus
          autocomplete="off"
          placeholder="show your query"
          v-model="query"
          @keyup.enter="submitQuery"
        />
        <div class="large-4 columns">
          <select v-model="timeSpan">
            <option value="3600" selected>Last 1 hour</option>
            <option value="7200">Last 2 hours</option>
            <option value="14400">Last 4 hours</option>
            <option value="28800">Last 8 hours</option>
            <option value="86400">Last 1 day</option>
            <option value="172800">Last 2 day</option>
            <option value="345600">Last 4 day</option>
            <option value="604800">Last 1 week</option>
            <option value="1209600">Last 2 week</option>
            <option value="2419200">Last 4 week</option>
          </select>
        </div>
        <div class="large-4 columns">
          <button class="send_query thin2" v-on:click="submitQuery">Query</button>
        </div>
        <div class="large-4 columns">
          <button
            class="secondary thin2"
            v-on:click="showApiKey"
            v-if="!showApiKeyForm"
          >Change API Key</button>
        </div>
      </div>
    </div>

    <div class="row" v-if="showApiKeyForm">
      <div class="columns large-2" style="text-align: right;">
        <h4>API Key</h4>
      </div>
      <div class="columns large-8">
        <input type="text" v-model="apiKey" />
      </div>
      <div class="large-2 columns">
        <button class="highlight thin" v-on:click="saveApiKey">Save</button>
      </div>
    </div>

    <div class="row" v-if="searchStatus !== null">
      <div class="large-12 columns query-status">
        <div class="alert-box radius">{{ searchStatus }}</div>
      </div>
    </div>

    <div class="row" v-if="errorMessage !== null">
      <div class="large-8 columns">
        <div class="alert-box alert">{{ errorMessage }}</div>
      </div>
      <div class="large-4 columns">
        <button class="alert-dark thin" v-on:click="clearError()">Dismiss</button>
      </div>
    </div>

    <!-- Log view --->
    <div class="row" v-if="logs.length > 0">
      <div class="large-9 push-3 columns">
        <!-- Pagenation (header) -->
        <div class="row" v-if="pages.length > 0">
          <ul class="pagination">
            <li v-for="p in pages" v-bind:class="{current: p.current}">
              <a href="#" v-on:click="changeSearchResultOffset(p.offset)">{{ p.index + 1 }}</a>
            </li>
          </ul>
        </div>

        <!-- Log view -->
        <div class="row">
          <table>
            <thead>
              <tr>
                <td>Meta</td>
                <td>Log</td>
              </tr>
            </thead>
            <tbody>
              <tr v-for="log in logs">
                <td class="log-meta-data">
                  <strong>{{ log.datetime }}</strong>
                  <span class="label" v-bind:style="log.labelStyle">{{ log.tag }}</span>
                </td>
                <td>
                  <table class="log-data-view">
                    <tbody>
                      <tr v-for="d in log.data">
                        <td class="log-field-column">{{ d.k }}</td>
                        <td>
                          <pre class="log-value-column" v-html="d.v"></pre>
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
        <div class="row" v-if="pages.length > 0">
          <ul class="pagination">
            <li v-for="p in pages" v-bind:class="{current: p.current}">
              <a href="#" v-on:click="changeSearchResultOffset(p.offset)">{{ p.index + 1 }}</a>
            </li>
          </ul>
        </div>
      </div>

      <div class="large-3 pull-9 columns">
        <div class="docs accordion metadata">
          <h3 class="metadata">Results</h3>
          <div class="content active">Elapsed time: {{ metadata.elapsed_seconds }} seconds</div>
          <div class="content active">Total: {{ metadata.total }} logs</div>
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
import sanitizeHTML from "sanitize-html";

var httpClient = axios.create({
  headers: { "x-api-key": localStorage.getItem("apiKey") }
});

const appData = {
  query: "",
  queryTerms: [],
  searchStatus: null,
  queryID: null,
  queryStart: null,
  apiKey: localStorage.getItem("apiKey"),
  showApiKeyForm: localStorage.getItem("apiKey") === null,
  logs: [],
  pages: [],
  metadata: {},
  timeSpan: 3600,
  errorMessage: null
};

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
  let values = [sanitizeHTML(raw)];
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

  if (appData.logs !== null) {
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
    .get(`/api/v1/search/` + appData.queryID + `/result?` + qs)
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

function getSearchResult(queryID, startDate) {
  const now = new Date();

  httpClient
    .get(`/api/v1/search/` + queryID + `/result`)
    .then(function(response) {
      console.log(response);

      switch (response.data.metadata.status) {
        case "SUCCEEDED":
          renderResult(response.data);
          break;

        case "RUNNING":
          const now = new Date();
          setTimeout(function() {
            appData.searchStatus =
              "Elapsed time: " +
              ((now.getTime() - appData.queryStart.getTime()) / 1000 +
                " seconds...");
            getSearchResult(queryID, startDate);
          }, 1000);
          break;

        default:
          showError("Fail request: " + response.data.metadata.status);
      }
    })
    .catch(showError);
}

function submitQuery(ev) {
  clearError();

  if (appData.apiKey === "") {
    showError("API key required");
    return;
  }

  console.log("submit...", ev);
  appData.logs = [];

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
  appData.queryStart = now;

  appData.queryTerms = appData.query.split(/\s+/).map(x => {
    return { term: x };
  });

  const body = {
    query: appData.queryTerms,
    start_dt: start_dt,
    end_dt: end_dt
  };

  httpClient
    .post(`/api/v1/search`, body)
    .then(function(response) {
      appData.searchStatus = "Start search";
      appData.queryID = response.data.query_id;
      setTimeout(function() {
        getSearchResult(response.data.query_id, now);
      }, 1000);
      console.log(response);
    })
    .catch(showError);
}

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
</script>
<style>
</style>
