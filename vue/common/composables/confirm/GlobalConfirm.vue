<!-- GlobalConfirm.vue -->
<script setup lang="ts">
import ZModal from '@common/ui/ZModal.vue'
import { useConfirmController } from './useConfirm'

const { state, handleOk, handleCancel } = useConfirmController()
</script>

<template>
  <ZModal :show="state.show" :closable="state.closable" :ok-text="state.okText" :ok-color="state.okColor"
    :cancel-text="state.cancelText" @ok="handleOk" @cancel="handleCancel">
    {{ state.message }}
  </ZModal>
</template>
