<template>
  <CContainer fluid>
    <CRow v-if="auth">
      <CCol md="12">
        <CCard>
          <CCardBody>
            <CRow>
              <CCol sm="12">
                <CFormInput
                  type="text"
                  autofocus
                  autocomplete="off"
                  placeholder="show your query"
                  v-on:update:value="query = $event"
                  v-bind:value="query"
                  v-on:keyup.native.enter="submitQuery"
                />
              </CCol>
            </CRow>
            <CRow>
              <CCol sm="2">
                <CFormSelect
                  v-on:update:value="spanMode = $event"
                  v-bind:value="spanMode"
                  :options="[
                { value: 'relative', label: 'Last' },
                { value: 'absolute', label: 'Between' }
                ]"
                />
              </CCol>

              <CCol sm="2" v-if="spanMode === 'relative'">
                <CFormSelect
                  v-on:update:value="timeSpan = $event"
                  :options="[
                  { value: '3600', label: '1 hour' },
                  { value: '7200', label: '2 hours' },
                  { value: '14400', label: '4 hours' },
                  { value: '28800', label: '8 hours' },
                  { value: '86400', label: '1 day' },
                  { value: '172800', label: '2 day' },
                  { value: '345600', label: '4 day' },
                  { value: '604800', label: '1 week' },
                  { value: '1209600', label: '2 week' },
                  { value: '2419200', label: '4 week' },
                ]"
                />
              </CCol>

              <CCol v-if="spanMode === 'absolute'">
                <input type="datetime-local" v-model="timeBegin" />
                To
                <input type="datetime-local" v-model="timeEnd" />
              </CCol>

              <CCol>
                <CButton
                  v-on:click="submitQuery"
                  class="m-2"
                  :key="'primary'"
                  :color="'primary'"
                >Query</CButton>
              </CCol>
            </CRow>
            <div class="row" v-if="errorMessage !== null">
              <div class="columns">
                <div class="msgbox alert">[Error] {{ errorMessage }}</div>
              </div>
              <div class="columns">
                <button class="alert-dark thin2" v-on:click="clearError()">Dismiss</button>
              </div>
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

function utcDateTime(unixtime) {
  const base = new Date(unixtime * 1000);
  const utc = new Date(
    base.getUTCFullYear(),
    base.getUTCMonth(),
    base.getUTCDate(),
    base.getUTCHours(),
    base.getUTCMinutes(),
    base.getUTCSeconds()
  );
  return strftime("%Y-%m-%dT%H:%M", utc);
}

const nowDatetime = new Date();

const appData = {
  query: "",
  timeSpan: 3600,
  timeBegin: utcDateTime(nowDatetime.getTime() / 1000 - 3600),
  timeEnd: utcDateTime(nowDatetime.getTime() / 1000),
  errorMessage: null,
  spanMode: "relative",
  auth: false
};

export default {
  data() {
    return appData;
  },
  methods: {
    clearError: clearError,
    submitQuery: submitQuery,
    setParameters: setParameters
  },
  mounted() {
    axios
      .get("/auth")
      .then(resp => {
        appData.auth = true;
      })
      .catch(err => {});

    this.setParameters();
  },
  watch: {
    $route(to, from) {
      if (to.matched[0].path === "/search/:search_id") {
        this.setParameters();
      }
    }
  }
};

function setParameters() {
  const searchID = this.$route.params.search_id;
  if (searchID === undefined) {
    return;
  }

  const url = `/api/v1/search/${searchID}`;
  axios
    .get(url)
    .then(response => {
      const meta = response.data.metadata;
      appData.query = meta.query.map(q => q.term).join(" ");
      appData.timeBegin = utcDateTime(meta.start_time);
      appData.timeEnd = utcDateTime(meta.end_time);
      appData.spanMode = "absolute";
    })
    .catch(showError);
}

function showError(err) {
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

  if (appData.query === "") {
    showError("No query");
    return;
  }

  appData.logs = [];

  const span = extractSpan();

  const terms = appData.query.split(/\s+/).map(x => {
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
      router.push("/search/" + response.data.search_id);
    })
    .catch(showError);
}
</script>
<style>
</style>
