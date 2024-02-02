<template>
  <CContainer fluid>
    <CRow v-if="errorMessage !== null">
      <CCol md="12">
        <CCard>
          <CCardBody>
            <CAlert color="danger" closeButton>{{ errorMessage }}</CAlert>
          </CCardBody>
        </CCard>
      </CCol>
    </CRow>

    <CRow v-if="progressMessage !== null">
      <CCol md="12">
        <CCard>
          <CCardBody>
            <CAlert color="primary">
              <CSpinner grow size="sm" />
              {{ progressMessage }}
            </CAlert>
          </CCardBody>
        </CCard>
      </CCol>
    </CRow>

    <CRow>
      <CCol md="2">
        <CCard>
          <CCardBody>
            <h3 class="metadata">Results</h3>
            <div v-if="metadata !== null">
              <div
                class="content"
              >Elapsed Time: {{ Math.floor(metadata.elapsed_seconds*1000) / 1000 }} seconds</div>
              <div class="content">
                Query Cost: ${{
                Math.floor((metadata.scanned_size * 5 / (1000*1000*1000*1000)) * 1000) / 1000 }}
                (Scanned {{
                Math.floor(metadata.scanned_size / (1000*1000)) / 1000 }} GB)
              </div>
              <div class="content">Total: {{ metadata.total }} logs</div>
              <div class="content">SubTotal: {{ metadata.sub_total }} logs</div>
              <div class="content">
                <div>Tags:</div>
                <div v-for="(tag, i) in metadata.tags" class="tag-selector">
                  <input type="checkbox" v-model="tags[tag]" v-on:change="changeSearchResultTags" />
                  {{ tag }}
                </div>
              </div>
            </div>
          </CCardBody>
        </CCard>
      </CCol>

      <!-- TODO: add "columns" class.
    Originally the <div> also should have "column", but layout will be broken if value is too long.
      I'm not good CSS writer for now-->
      <CCol>
        <CCard v-if="chartDataSets.length > 0">
          <CChartBar
            style="width:99%; height: 200px"
            :datasets="chartDataSets"
            :labels="chartLabels"
            :options="chartOptions"
          />
        </CCard>

        <CCard>
          <CCardHeader>
            <CFormInput
              type="text"
              autocomplete="off"
              placeholder="jq filter"
              v-on:update:value="query = $event"
              v-on:keyup.native.enter="renewJqQuery"
              v-bind:value="query"
            />
            <div class="subrow msgbox sysmsg" v-if="systemMessage !== null">{{ systemMessage }}</div>
          </CCardHeader>

          <CCardBody v-if="metadata !== null">
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
                        <div>
                          <strong>{{ log.datetime }}</strong>
                        </div>
                        <div>
                          <span class="label" v-bind:style="log.labelStyle">{{ log.tag }}</span>
                        </div>
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
          </CCardBody>
        </CCard>
      </CCol>
    </CRow>
  </CContainer>
</template>
<script>
import axios from "axios";
import strftime from "strftime";
import querystring from "querystring";
import SHA1 from "crypto-js/sha1";
import escapeHTML from "escape-html";
import { prototype } from "events";

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
  tags: [],
  searchStatus: null,
  logs: [],
  pages: [],
  metadata: null,
  errorMessage: null,
  systemMessage: null,
  progressMessage: null,

  chartLabels: ["January", "February", "March", "April", "May", "June", "July"],
  chartOptions: {
    // animation: false,
    maintainAspectRatio: false,
    onClick: onClickChart,
    scales: {
      xAxes: [{ stacked: true }],
      yAxes: [{ stacked: true, ticks: { min: 0 } }]
    }
  },
  chartDataSets: [],
  // lastQueryString will be set this.$route.query
  lastQueryString: null,
  search_id: null
};

setTimeout(() => {}, 2000);

export default {
  data() {
    return appData;
  },
  methods: {
    changeSearchResultOffset: changeSearchResultOffset,
    changeSearchResultTags: changeSearchResultTags,
    showSearch: showSearch,
    renewJqQuery: renewJqQuery,
    clearError: clearError,
    onClickChart: onClickChart
  },
  mounted() {
    this.showSearch();
  },
  watch: {
    $route(to, from) {
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

function toColor(text) {
  return (
    "#" +
    SHA1(text)
      .toString()
      .substring(0, 6)
  );
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
    currentPage !== undefined ? [currentPage] : [],
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
  appData.metadata = data.metadata;
  appData.searchStatus = null;

  if (data.logs === null || data.logs.length === 0) {
    appData.systemMessage = "No log found";
    appData.logs = [];
    return;
  }

  const enabledTags = {};
  if (appData.lastQueryString.tags === undefined) {
    data.metadata.tags.forEach(tag => {
      enabledTags[tag] = true;
    });
  } else {
    appData.lastQueryString.tags.split(",").forEach(tag => {
      enabledTags[tag] = true;
    });
  }
  data.metadata.tags.forEach(t => {
    appData.tags[t] = enabledTags[t] ? true : false;
  });

  appData.pages = buildPagenationIndices(data.metadata);

  appData.logs = data.logs.map(x => {
    const bgColor = toColor(x.tag);

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
  appData.systemMessage = null;

  const qs = Object.assign(appData.lastQueryString, newQuery);
  Object.keys(newQuery).forEach(k => {
    if (newQuery[k] === null) {
      delete qs[k];
    }
  });

  const url =
    `/search/${appData.search_id}` +
    (Object.keys(qs).length > 0 ? "?" + querystring.stringify(qs) : "");
  router.push(url);
}

function renewJqQuery() {
  if (appData.query === "") {
    renewSearchResult(this.$router, { query: null });
  } else {
    renewSearchResult(this.$router, { query: appData.query });
  }
}

function changeSearchResultOffset(offset, current) {
  if (current || offset === undefined) {
    return;
  }

  renewSearchResult(this.$router, { offset: offset });
}

function changeSearchResultTags(args) {
  if (Object.keys(appData.tags).every(v => appData.tags[v])) {
    renewSearchResult(this.$router, { tags: null });
  } else {
    const tags = Object.keys(appData.tags)
      .map(t => (appData.tags[t] ? t : null))
      .filter(x => x !== null);
    renewSearchResult(this.$router, { tags: tags.join(",") });
  }
}

function getSearchLogs(searchID, qs) {
  const url =
    `/api/v1/search/${searchID}/logs` +
    (Object.keys(qs).length > 0 ? "?" + querystring.stringify(qs) : "");

  appData.progressMessage = null;
  axios
    .get(url)
    .then(response => {
      renderResult(response.data);
    })
    .catch(err => {
      if (err.response && err.response.data.message) {
        showError(err.response.data.message);
      }
    });
}

function renderChart(data) {
  appData.chartLabels = data.labels;
  appData.chartDataSets = Object.keys(data.timeseries).map(k => {
    return {
      data: data.timeseries[k],
      barPercentage: 0.99,
      categoryPercentage: 0.99,
      backgroundColor: toColor(k),
      label: k
    };
  });
}

function getSearchTimeSeries(searchID) {
  const url = `/api/v1/search/${searchID}/timeseries`;

  axios
    .get(url)
    .then(response => {
      renderChart(response.data);
    })
    .catch(err);
      if (err.response && err.response.data.message) {
        showError(err.response.data.message);
      };
}

function getSearchResult(searchID, qs, n = 1) {
  clearError();
  appData.logs = [];

  const now = new Date();

  const url = `/api/v1/search/${searchID}`;

  axios
    .get(url)
    .then(function(response) {
      switch (response.data.metadata.status) {
        case "SUCCEEDED":
          appData.metadata = response.data.metadata;
          getSearchTimeSeries(searchID);
          getSearchLogs(searchID, qs);
          break;

        case "RUNNING":
          const now = new Date();
          appData.progressMessage =
            "Elapsed time: " +
            Math.floor(response.data.metadata.elapsed_seconds * 100) / 100 +
            " seconds";

          const waitTime = ((n ^ 2) / 16 + 1) * 1000;
          setTimeout(function() {
            getSearchResult(searchID, qs, n + 1);
          }, waitTime);
          break;

        default:
          showError("Fail request: " + response.data.metadata.status);
      }
    })
    .catch(err => {
      if (err.response && err.response.data.message) {
        showError(err.response.data.message);
      }
    });
}

function showSearch() {
  appData.search_id = this.$route.params.search_id;
  appData.query = this.$route.query.query;
  appData.lastQueryString = Object.assign({}, this.$route.query);
  appData.metadata = null;
  getSearchResult(this.$route.params.search_id, this.$route.query);
}
</script>
<style>
</style>
