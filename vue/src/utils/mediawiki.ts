import http from '@/utils/http'

export default async function titleExist(title: string): Promise<boolean> {
  try {
    const { data } = await http.get('/w/api.php', { params: { action: 'query', format: 'json', titles: title } })
    const pages = data?.query?.pages;
    return pages && Object.keys(pages).some(key => key !== '-1');
  } catch (error) {
    console.error('Error fetching title existence:', error);
    return false;
  }
}
