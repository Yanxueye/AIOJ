import http from './index'

export const tagApi = {
  // Get all tags grouped by category
  getList: () => http.get('/tags'),
  // Get flat list of tag names (for AI prompt injection)
  getNames: () => http.get('/tags/names')
}
