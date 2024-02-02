<template>
  <CHeader fixed light>
    <CHeaderNav class="d-md-down-none mr-auto">
      <CNavItem class="px-3">
        <CNavLink to="/">Strix</CNavLink>
      </CNavItem>
    </CHeaderNav>
    <CHeaderNav class="mr-4">
      <CNavItem class="d-md-down-none mx-2" v-if="user === null">
        <CButton color="primary" class="m-2" v-on:click="moveToLoginPage">
          <!--    <a href="/auth/google">Login</a>-->
          Login
        </CButton>
      </CNavItem>

      <CDropdown
        v-else
        inNav
        class="c-header-nav-items"
        placement="bottom-end"
        add-menu-classes="pt-0"
      >
        <template #toggler>
          <CNavLink>
            <div class="c-avatar">
              <img :src="user.image" class="c-avatar-img" />
            </div>
          </CNavLink>
        </template>

        <CDropdownHeader tag="div" class="text-center" color="light">
          <strong>{{ user.user }}</strong>
        </CDropdownHeader>
        <CDropdownItem href="/auth/logout">Logout</CDropdownItem>
      </CDropdown>
    </CHeaderNav>
  </CHeader>
</template>

<script>
import axios from "axios";

const appData = {
  user: null
};

export default {
  data() {
    return appData;
  },
  methods: {
    moveToLoginPage: function() {
      window.location = "/auth/google";
    }
  },
  mounted() {
    axios
      .get("/auth")
      .then(resp => {
        console.log("Auth OK:", resp);
        appData.user = resp.data.user;
      })
      .catch(err => {
        console.log("auth NG", err);
      });
  }
};
</script>
<style>
</style>
