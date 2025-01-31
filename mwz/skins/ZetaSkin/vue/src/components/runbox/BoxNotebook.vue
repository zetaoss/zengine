<script lang='ts'>
import { inject, computed } from 'vue';
import { type Job } from './types';

export default {
  setup() {
    const job = inject<Job>('job');
    const outs = computed(() => job?.response?.outs ? JSON.parse(JSON.stringify(job.response.outs)) : []);

    const ansi2html = (s: string) => {
      let isOpen = false;
      return s
        .replace(/</g, '&lt;')
        .replace(/\x1b\[(\d+;)?(\d+)m/g, (_, _1, code) => {
          const closeTag = isOpen ? '</span>' : '';
          isOpen = !!code;
          return closeTag + (code ? `<span class="ansi${code}">` : '');
        }) + (isOpen ? '</span>' : '');
    };

    return { outs, ansi2html };
  },
};
</script>

<template>
  <div class="border p-2">
    <slot />
  </div>
  <div class="out">
    <div v-if="outs.length" class="border">
      <div v-for="(out, i) in outs" :key="i" class="outputs">
        <div v-for="(r, j) in out" :key="j" class="output p-2" :class="r.output_type">

          <template v-if="r.output_type === 'stream' && r.text">
            <div v-for="(text, k) in r.text" :key="k">{{ text }}</div>
          </template>

          <template v-else-if="r.output_type === 'display_data' && r.data">
            <div v-if="r.data['image/png']">
              <img :src="'data:image/png;base64,' + r.data['image/png']" />
            </div>
          </template>

          <template v-else-if="r.output_type === 'execute_result' && r.data">
            <div v-if="r.data['text/html']">
              <div v-html="r.data['text/html'].join('')"></div>
            </div>
            <div v-else-if="r.data['text/plain']">
              <pre>{{ r.data['text/plain'].join('') }}</pre>
            </div>
          </template>

          <template v-else-if="r.output_type === 'error'">
            <pre v-html="ansi2html(r.traceback?.join('\n'))" />
          </template>

        </div>
      </div>
    </div>
  </div>
</template>

<style>
.out {
  font-family: "Roboto", "Noto", sans-serif;
  @apply text-sm;

  img {
    @apply bg-white;
  }

  .dataframe {
    @apply border;

    th {
      @apply px-2;
    }

    tbody {
      tr {
        text-align: right;

        &:nth-child(odd) {
          background: #8884;
        }
      }
    }
  }
}

.output.error {
  @apply bg-red-100 dark:bg-red-950/80;
}

.ansi30 {
  @apply text-black dark:text-gray-500;
}

.ansi31 {
  @apply text-red-600 dark:text-red-400;
}

.ansi32 {
  @apply text-green-600 dark:text-green-400;
}

.ansi33 {
  @apply text-yellow-600 dark:text-yellow-400;
}

.ansi34 {
  @apply text-blue-600 dark:text-blue-400;
}

.ansi35 {
  @apply text-purple-600 dark:text-purple-400;
}

.ansi36 {
  @apply text-cyan-600 dark:text-cyan-400;
}

.ansi37 {
  @apply text-gray-600 dark:text-gray-400;
}

.ansi90 {
  @apply text-gray-500 dark:text-gray-600;
}

.ansi91 {
  @apply text-red-400 dark:text-red-300;
}

.ansi92 {
  @apply text-green-400 dark:text-green-300;
}

.ansi93 {
  @apply text-yellow-400 dark:text-yellow-300;
}

.ansi94 {
  @apply text-blue-400 dark:text-blue-300;
}

.ansi95 {
  @apply text-purple-400 dark:text-purple-300;
}

.ansi96 {
  @apply text-cyan-400 dark:text-cyan-300;
}

.ansi97 {
  @apply text-gray-400 dark:text-gray-200;
}
</style>
