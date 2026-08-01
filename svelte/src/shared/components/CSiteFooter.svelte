<svelte:options customElement={{ tag: 'c-site-footer', shadow: 'none' }} />

<script lang="ts">
  import { onMount } from 'svelte'

  import CCookieSettings from '$shared/components/CCookieSettings.svelte'
  import { getTrackingState } from '$shared/stores/trackingStore'
  import versionData from '$shared/version.json'

  let consentOpen = false
  let showCookieSettings = false

  onMount(() => {
    const state = getTrackingState()
    const isStrict = state.policy === 'strict'
    showCookieSettings = isStrict
    consentOpen = isStrict && !state.hasConsentCookie
  })
</script>

<footer id="footer" class="border-t border-border bg-card text-card-foreground p-8">
  <div class="mx-auto max-w-8xl">
    <section class="flex flex-col gap-6 pb-4 md:flex-row md:items-end md:justify-between">
      <div>
        <img class="h-8" src="/zeta2.svg" />
        <div class="text-xl font-bold text-foreground pt-4 pb-2">제타위키</div>
        <div class="text-sm text-a-gray-600" title="{versionData.buildDate}">세상의 각주, 기록하고 연결합니다.</div>
      </div>
      <nav>
        <ul class="list-none pl-0 flex gap-3">
          <li>
            <a
              href="https://creativecommons.org/licenses/by-sa/4.0/"
              target="_blank"
              title="Creative Commons BY-SA 4.0"
              class="bg-white flex h-11 rounded border p-2"
            >
              <img
                src="https://upload.wikimedia.org/wikipedia/commons/e/e5/CC_BY-SA_icon.svg"
                class="h-full w-auto"
                alt="Creative Commons BY-SA 4.0"
              />
            </a>
          </li>
          <li>
            <a
              href="https://www.mediawiki.org/"
              target="_blank"
              title="Powered by MediaWiki"
              class="bg-white flex h-11 rounded border p-2"
            >
              <img
                src="https://mediawiki.org/w/resources/assets/poweredby_mediawiki.svg"
                class="h-full w-auto"
                alt="Powered by MediaWiki"
              />
            </a>
          </li>
        </ul>
      </nav>
    </section>

    <!-- Bottom Section -->
    <section class="flex flex-col gap-4 border-t border-border/60 pt-4 md:flex-row md:items-center md:justify-between">
      <ul class="list-none pl-0 flex flex-wrap items-center gap-x-4 gap-y-2 text-sm text-muted-foreground">
        <li>
          <a href="/wiki/제타위키:개인정보처리방침">개인정보처리방침</a>
        </li>
        <li>
          <a href="/wiki/제타위키:면책_조항">면책조항</a>
        </li>
        <li>
          <a href="/wiki/제타위키">제타위키 소개</a>
        </li>
        {#if showCookieSettings}
          <li>
            <button class="cursor-pointer transition-colors hover:text-foreground" type="button" on:click={() => (consentOpen = true)}>
              쿠키설정
            </button>
          </li>
        {/if}
      </ul>
      <nav id="footer-icons" class="noprint" aria-label="Footer Icons"></nav>
    </section>
  </div>

  {#if showCookieSettings}
    <CCookieSettings open={consentOpen} onClose={() => (consentOpen = false)} />
  {/if}
</footer>
