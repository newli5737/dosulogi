import { http } from '@/shared/api/http'
import type { Contact, ContactInput } from '../model/types'

export const contactApi = {
  list: (customerId: string): Promise<{ data: Contact[] }> =>
    http(`/api/v1/customers/${customerId}/contacts`),
  create: (customerId: string, body: ContactInput): Promise<{ data: Contact }> =>
    http(`/api/v1/customers/${customerId}/contacts`, { method: 'POST', body }),
  update: (customerId: string, contactId: string, body: ContactInput): Promise<{ data: Contact }> =>
    http(`/api/v1/customers/${customerId}/contacts/${contactId}`, { method: 'PUT', body }),
  remove: (customerId: string, contactId: string): Promise<void> =>
    http(`/api/v1/customers/${customerId}/contacts/${contactId}`, { method: 'DELETE' }),
}
