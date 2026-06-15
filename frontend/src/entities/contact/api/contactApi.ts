import { http } from '@/shared/api/http'
import type { Contact, ContactInput } from '../model/types'

export const contactApi = {
  list: (token: string, customerId: string): Promise<{ data: Contact[] }> =>
    http(`/api/v1/customers/${customerId}/contacts`, { token }),
  create: (token: string, customerId: string, body: ContactInput): Promise<{ data: Contact }> =>
    http(`/api/v1/customers/${customerId}/contacts`, { token, method: 'POST', body }),
  update: (token: string, customerId: string, contactId: string, body: ContactInput): Promise<{ data: Contact }> =>
    http(`/api/v1/customers/${customerId}/contacts/${contactId}`, { token, method: 'PUT', body }),
  remove: (token: string, customerId: string, contactId: string): Promise<void> =>
    http(`/api/v1/customers/${customerId}/contacts/${contactId}`, { token, method: 'DELETE' }),
}
