<template>
  <CContainer fluid>
    <CRow>
      <CCol md="12">
        <CCard>
          <CCardBody>
            <CRow>
              <CCol sm="12">
                <CInput
                  type="text"
                  autofocus
                  autocomplete="off"
                  placeholder="show your query"
                  v-on:update:value="query = $event"
                  v-on:keyup.native.enter="submitQuery"
                />
              </CCol>
            </CRow>
            <CRow>
              <CCol sm="2">
                <CSelect
                  v-on:update:value="spanMode = $event"
                  :options="[
                { value: 'relative', label: 'Last' },
                { value: 'absolute', label: 'Between' }
                ]"
                />
              </CCol>

              <CCol sm="2" v-if="spanMode === 'relative'">
                <CSelect
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

                <div class="columns">
                  <button class="secondary thin2" v-if="!authenticated">
                    <a href="/auth/google">Google Login</a>
                  </button>
                </div>
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
      router.push("/search/" + response.data.query_id);
    })
    .catch(showError);
}
</script>
<style>
</style>
