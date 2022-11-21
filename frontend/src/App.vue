<script setup lang="ts">
import { onMounted, ref, computed } from "@vue/runtime-core";
import debounce from "./utils/debounce";

type StudioInfo = {
  id: number;
  magiclineId: number;
  name: string;
  uuid: string;
};

const isLoading = ref(false);
const searchTerm = ref("");

const studios = ref<Array<StudioInfo>>([]);
const filteredStudios = ref<Array<StudioInfo>>([]);

const loadStudios = async () => {
  isLoading.value = true;
  const res = await fetch("/api/studios");
  const data = await res.json();
  studios.value = data.content as Array<StudioInfo>;
  isLoading.value = false;
};

const filter = debounce(() => {
  if (!searchTerm.value || searchTerm.value === "") {
    filteredStudios.value = studios.value;
  } else {
    filteredStudios.value = studios.value.filter((s) =>
      s.name.toLowerCase().includes(searchTerm.value.toLowerCase())
    );
  }
}, 200);
onMounted(async () => {
  await loadStudios();
  filteredStudios.value = studios.value;
});
</script>

<template>
  <v-app id="inspire">
    <v-app-bar app>
      <v-toolbar-title>FitX Studio Finder</v-toolbar-title>
    </v-app-bar>

    <v-main>
      <v-container>
        <v-row>
          <v-col cols="12">
            <v-text-field
              v-model="searchTerm"
              @keyup="filter"
              variant="outlined"
              label="Suche"
            ></v-text-field>
          </v-col>
        </v-row>
        <v-slide-y-transition mode="out-in">
          <v-row v-if="isLoading">
            <v-col cols="12">
              <v-progress-linear indeterminate />
            </v-col>
          </v-row>
          <v-row v-else>
            <v-col
              v-for="s in filteredStudios"
              :key="s.id"
              cols="12"
              sm="6"
              lg="4"
            >
              <v-card min-height="200">
                <v-card-title>
                  {{ s.name }}
                </v-card-title>
                <v-card-text>
                  <v-table>
                    <tbody>
                      <tr>
                        <td>Studio-ID:</td>
                        <td>{{ s.id }}</td>
                      </tr>
                      <tr>
                        <td>Studio-UUID:</td>
                        <td>{{ s.uuid }}</td>
                      </tr>
                      <tr>
                        <td>Aktuelle Magicline:</td>
                        <td>{{ s.magiclineId }}</td>
                      </tr>
                    </tbody>
                  </v-table>
                </v-card-text>
                <v-card-actions>
                  <v-spacer />
                  <v-btn
                    target="_blank"
                    :href="`https://mein.fitx.de/studio/${s.uuid}`"
                  >
                    Open
                  </v-btn>
                </v-card-actions>
              </v-card>
            </v-col>
          </v-row>
        </v-slide-y-transition>
      </v-container>
    </v-main>
  </v-app>
</template>
