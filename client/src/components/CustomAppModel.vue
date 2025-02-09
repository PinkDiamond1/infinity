<template>
  <q-card>
    <q-card-section>
      <q-btn
        v-if="!model.single"
        unelevated
        color="primary"
        @click="openCreateDialog"
        >New {{ model.display || model.name }}</q-btn
      >
      <q-btn
        v-else
        unelevated
        color="primary"
        @click="openUpdateDialog('single')"
        >Edit {{ model.display || model.name }}</q-btn
      >
    </q-card-section>
    <q-card-section>
      <div v-if="!model.single" class="row items-center no-wrap q-mb-md">
        <div class="col">
          <h5 class="text-subtitle1 q-my-none">
            {{
              model.plural ||
              (model.display ? `${model.display}s` : `${model.name}s`)
            }}
          </h5>
        </div>
        <div class="col-4">
          <div class="row items-center q-gutter-sm float-right on-right">
            <div class="col-2">
              <q-btn
                round
                size="sm"
                color="deep-orange"
                icon="filter_alt"
                @click="openFilterDialog"
              />
            </div>
            <div
              class="col q-mx-md"
              :style="{
                fontSize: '80%',
                whiteSpace: 'pre-wrap',
                wordBreak: 'break-all'
              }"
            >
              <code>{{ filterString }}</code>
            </div>
          </div>
        </div>
      </div>

      <q-table
        v-if="!model.single"
        v-model:pagination="pagination"
        dense
        flat
        binary-state-sort
        :columns="columns"
        :filter="filters"
        :filter-method="filterMethod"
        :rows="items[model.name]"
        row-key="key"
      >
        <template #body="props">
          <q-tr :props="props">
            <q-td auto-width class="text-center">
              <code :title="formatDate(props.row.created_at, true)">
                {{ props.row.key }}
              </code>
            </q-td>
            <q-td
              v-for="field in model.fields"
              :key="field.name"
              auto-width
              class="text-center"
            >
              <AppPropertyDisplay
                :field="field"
                :data="props.row.value"
                :items-map="itemsMap"
              />
            </q-td>
            <q-td auto-width>
              <q-btn
                flat
                dense
                size="xs"
                icon="edit"
                color="light-blue"
                @click="openUpdateDialog(props.row.key)"
              ></q-btn>
              <q-btn
                flat
                dense
                size="xs"
                icon="cancel"
                color="pink"
                @click="deleteItem(props.row.key)"
              ></q-btn>
            </q-td>
          </q-tr>
        </template>
      </q-table>
      <q-list v-else dense>
        <q-item v-for="field in model.fields" :key="field.name">
          <q-item-section class="col">
            <q-item-label>{{ fieldLabel(field) }}</q-item-label>
          </q-item-section>
          <q-item-section class="col" style="font-size: 13px">
            <q-item-label>
              <AppPropertyDisplay
                single
                :field="field"
                :data="singleItem.value"
                :items-map="itemsMap"
              />
            </q-item-label>
          </q-item-section>
        </q-item>
      </q-list>
    </q-card-section>
  </q-card>

  <q-dialog v-model="dialog.show" @hide="closeDialog">
    <q-card class="q-pa-lg lnbits__dialog-card">
      <!-- ITEM EDITING MODAL -->
      <q-form v-if="dialog.item" class="q-gutter-md" @submit="saveItem">
        <div class="text-h6">
          Editing
          <span v-if="dialog.item.key === 'single'">
            {{ model.display || model.name }}
          </span>
          <code v-else>{{ dialog.item.key }}</code>
        </div>
        <template
          v-for="field in model.fields.filter(f => !f.computed)"
          :key="field.name"
        >
          <AppPropertyEdit
            v-model:data="dialog.item.value"
            :field="field"
            :items="items"
          />
        </template>
        <div class="row q-mt-lg">
          <q-btn v-if="dialog.item.key" unelevated color="primary" type="submit"
            >Update {{ model.name }}</q-btn
          >
          <q-btn
            v-else
            unelevated
            color="primary"
            :disabled="isFormSubmitDisabled"
            type="submit"
            >Create {{ model.name }}</q-btn
          >
          <q-btn v-close-popup flat color="grey" class="q-ml-auto"
            >Cancel</q-btn
          >
        </div>
      </q-form>
      <!-- END ITEM EDITING MODAL -->

      <!-- FILTERS MODAL -->
      <q-form v-if="dialog.filters" class="q-gutter-md">
        <div class="text-h6">Filters</div>
        <template v-for="(filter, f) in dialog.filters" :key="f">
          <div class="row">
            <div class="col">
              <q-select
                v-model="filter.field"
                filled
                dense
                emit-value
                map-options
                clearable
                :options="
                  model.fields
                    .filter(field => field.type !== 'ref')
                    .map(field => ({
                      value: field.name,
                      label: fieldLabel(field)
                    }))
                "
                label="Field"
              />
            </div>

            <div class="col-2 q-mx-sm">
              <q-select
                v-model="filter.op"
                dense
                filled
                :options="['=', '!=', '~', '<', '>', '<=', '>=']"
              />
            </div>

            <div class="col">
              <template
                v-if="filter.field && fieldsMap && fieldsMap[filter.field]"
              >
                <q-input
                  v-if="
                    fieldsMap[filter.field].type === 'string' ||
                    fieldsMap[filter.field].type === 'url'
                  "
                  v-model.trim="filter.value"
                  filled
                  dense
                  :type="
                    fieldsMap[filter.field].type === 'url' ? 'url' : 'text'
                  "
                  label="Value"
                />
                <q-input
                  v-if="fieldsMap[filter.field].type === 'number'"
                  v-model.number="filter.value"
                  filled
                  dense
                  type="number"
                  label="Value"
                />
                <q-input
                  v-if="fieldsMap[filter.field].type === 'msatoshi'"
                  filled
                  dense
                  type="text"
                  suffix="satoshis"
                  label="Value"
                  :model-value="filter.value > 0 ? filter.value / 1000 : ''"
                  @update:model-value="
                    filter.value = (parseInt($event) || 0) * 1000
                  "
                />
                <q-toggle
                  v-if="fieldsMap[filter.field].type === 'boolean'"
                  v-model="filter.value"
                  label="Value"
                  :indeterminate-value="'INDETERMINATE'"
                />
              </template>
            </div>
          </div>
        </template>
      </q-form>
      <!-- END FILTERS MODAL -->
    </q-card>
  </q-dialog>
</template>

<script>
import {setAppItem, addAppItem, delAppItem} from '../api'
import {paramDefaults, formatDate, fieldLabel, notifyError} from '../helpers'

export default {
  props: {
    model: {
      type: Object,
      required: true
    },
    items: {
      type: Object,
      required: true
    },
    itemsMap: {
      type: Object,
      required: true
    }
  },

  data() {
    return {
      dialog: {
        show: false,
        item: null, // the same dialog object is used for item adding/editing
        filter: null // and for filter adding/editing
      },
      filters: null,
      pagination: {
        rowsPerPage: 15,
        sortBy: 'created_at',
        descending: false,
        ...this.model.defaultSort
      }
    }
  },

  computed: {
    filterString() {
      return this.filters
        ? this.filters
            .map(({field, op, value}) => `${field} ${op} ${value}`)
            .join('; ')
        : null
    },

    singleItem() {
      return (
        this.items[this.model.name].find(item => item.key === 'single') || {
          wallet: this.$store.state.wallet.id,
          model: this.model.name,
          key: 'single',
          value: paramDefaults(this.model.fields)
        }
      )
    },

    fieldsMap() {
      return Object.fromEntries(
        this.model.fields.map(field => [field.name, field])
      )
    },

    columns() {
      const headerStyle = 'text-align: center;'

      return [
        {name: 'key', label: 'Key', field() {}, sortable: true, headerStyle},
        ...this.model.fields.map(field => ({
          name: field.name,
          label: fieldLabel(field),
          field: row => this.getSortableFieldValue(row, field),
          sortable: true,
          headerStyle: 'font-size: 110%;' + headerStyle
        })),
        {name: '_controls', label: '', field() {}, headerStyle}
      ]
    },

    isFormSubmitDisabled() {
      return (
        this.dialog.show &&
        this.model.fields
          .filter(field => field.required)
          .filter(
            field =>
              this.dialog.item.value[field.name] === undefined ||
              this.dialog.item.value[field.name] === ''
          ).length > 0
      )
    }
  },

  mounted() {
    this.filters = this.model.defaultFilters
  },

  methods: {
    json: v => JSON.stringify(v, null, 2),

    formatDate,
    fieldLabel,

    openFilterDialog() {
      this.dialog = {
        filters: [...(this.filters || []), {field: null, op: '=', value: ''}],
        show: true,
        item: null
      }
    },

    openCreateDialog() {
      this.dialog = {
        item: {
          wallet: this.$store.state.wallet.id,
          model: this.model.name,
          value: paramDefaults(this.model.fields)
        },
        show: true,
        filter: null
      }
    },

    openUpdateDialog(key) {
      var item = this.items[this.model.name].find(item => item.key === key)
      if (!item && key === 'single') {
        item = {
          wallet: this.$store.state.wallet.id,
          model: this.model.name,
          key,
          value: paramDefaults(this.model.fields)
        }
      }

      item = {...item, value: {...item.value}}
      this.model.fields
        .filter(field => field.computed)
        .forEach(f => {
          delete item.value[f.name]
        })
      this.dialog = {item, show: true, filter: null}
    },

    closeDialog() {
      if (this.dialog.filters)
        this.filters = this.dialog.filters.filter(
          ({field, op, value}) => field && op && value
        )

      this.dialog.show = false
    },

    async saveItem() {
      try {
        if (this.dialog.item.key) {
          await setAppItem(
            this.$store.state.app.url,
            this.model.name,
            this.dialog.item.key,
            this.dialog.item.value
          )
        } else {
          await addAppItem(
            this.$store.state.app.url,
            this.model.name,
            this.dialog.item.value
          )
        }

        this.$q.notify({
          message: `${this.model.display || this.model.name} saved.`,
          type: 'positive',
          timeout: 3500
        })

        this.closeDialog()
      } catch (err) {
        notifyError(err)
      }
    },

    deleteItem(key) {
      this.$q
        .dialog({
          message: 'Are you sure you want to delete this item?',
          ok: {
            flat: true,
            color: 'orange'
          },
          cancel: {
            flat: true,
            color: 'grey'
          }
        })
        .onOk(async () => {
          try {
            await delAppItem(this.$store.state.app.url, this.model.name, key)
            this.$q.notify({
              message: `${this.model.display || this.model.name} deleted.`,
              type: 'info',
              timeout: 2500
            })
          } catch (err) {
            notifyError(err)
          }
        })
    },

    getSortableFieldValue(item, field) {
      switch (field.type) {
        case 'currency':
          return item.value[field.name].amount
        default:
          return item.value[field.name]
      }
    },

    filterMethod(rows, filters) {
      return rows.filter(({value: item}) => {
        for (let i = 0; i < filters.length; i++) {
          let {field, op, value} = filters[i]

          switch (op) {
            case '=': {
              if (item[field] !== value) return false
              break
            }
            case '!=': {
              if (item[field] === value) return false
              break
            }
            case '~': {
              if (!item[field] || item[field].indexOf) return false
              if (item[field].indexOf(value) === -1) return false
              break
            }
            case '>': {
              if (item[field] <= value) return false
              break
            }
            case '>=': {
              if (item[field] < value) return false
              break
            }
            case '<': {
              if (item[field] >= value) return false
              break
            }
            case '<=': {
              if (item[field] > value) return false
              break
            }
          }
        }

        return true
      })
    }
  }
}
</script>
