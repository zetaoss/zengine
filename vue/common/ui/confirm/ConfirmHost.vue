<!-- ConfirmHost.vue -->
<script setup lang="ts">
import ZModal from '@common/ui/ZModal.vue'

import { confirmState, handleConfirmCancel, handleConfirmOk } from './confirm'

</script>

<template>
  <ZModal :show="confirmState.show" :closable="confirmState.closable" :ok-text="confirmState.okText"
    :ok-color="confirmState.okColor" :cancel-text="confirmState.cancelText" @ok="handleConfirmOk"
    @cancel="handleConfirmCancel">
    {{ confirmState.message }}
  </ZModal>
</template>
