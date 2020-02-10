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
        <button class="secondary thin2" v-if="!authenticated">
          <a href="/auth/google">Google Login</a>
        </button>
      </div>
    </div>

    <div class="row" v-if="errorMessage !== null">
      <div class="columns">
        <div class="msgbox alert">[Error] {{ errorMessage }}</div>
      </div>
      <div class="columns">
        <button class="alert-dark thin2" v-on:click="clearError()">Dismiss</button>
      </div>
    </div>
  </div>
</template>

<script>
import axios from "axios";
import strftime from "strftime";

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
  timeSpan: 3600,
  timeBegin: strftime("%Y-%m-%dT%H:%M", utcDatetime),
  timeEnd: strftime("%Y-%m-%dT%H:%M", utcDatetime),
  errorMessage: null,
  spanMode: "relative",
  authenticated: false
};

export default {
  data() {
    return appData;
  },
  mounted() {
    axios
      .get("/auth")
      .then(resp => {
        appData.authenticated = true;
      })
      .catch(err => {
        console.log("auth NG", err);
      });
  },
  methods: {
    clearError: clearError,
    submitQuery: submitQuery
  }
};

function showError(err) {
  console.log(err);
  if (err.response) {
    console.log("error response: ", err.response);
  }

  appData.errorMessage = err;
}

function clearError() {
  appData.errorMessage = null;
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

  terms = appData.query.split(/\s+/).map(x => {
    return { term: x };
  });

  const body = {
    query: terms,
    start_dt: span.start,
    end_dt: span.end
  };
  console.log("body =>", body);

  const router = this.$router;
  axios
    .post(`/api/v1/search`, body)
    .then(response => {
      console.log(response);
      router.push("/search/" + response.data.query_id);
    })
    .catch(showError);
}
</script>
<style>
</style>
