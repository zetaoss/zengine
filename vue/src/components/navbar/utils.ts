import type Item from '@common/components/navbar/types'
import type { Avatar } from '@common/components/avatar/avatar'

export default function getUserMenuItems(avatar: Avatar) {
  const isLoggedIn = (avatar?.id ?? 0) > 0;

  const items: Item[] = isLoggedIn
    ? [
      { text: '사용자 문서', href: '/wiki/특수:내사용자문서' },
      { text: '사용자 토론', href: '/wiki/특수:내사용자토론' },
      { text: '환경 설정', href: '/wiki/특수:환경설정' },
      { text: '주시문서 목록', href: '/wiki/특수:주시문서목록' },
      { text: '기여', href: '/wiki/특수:내기여' },
      { text: '업로드', href: '#' }, // 빈 문자열 대신 # 사용
      { text: '특수문서', href: '#' }, // 빈 문자열 대신 # 사용
      { text: '로그아웃', href: '/logout' },
    ]
    : [
      { text: '토론', href: '/wiki/특수:내사용자토론' },
      { text: '기여', href: '/wiki/특수:내기여' },
      { text: '계정 생성', href: '/wiki/특수:계정만들기' },
      { text: '로그인', href: '/login' },
    ]

  return items
}
