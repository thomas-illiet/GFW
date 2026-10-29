import { describe, expect, it } from 'vitest'

import { toCsv } from '../app/utils/export'
import { dateTime, number, percent, relativeTime } from '../app/utils/format'
import {
  instancePresentation,
  localInstanceEnvironment,
} from '../app/utils/instance'

describe('dashboard formatters', () => {
  it('formats dashboard values', () => {
    expect(number(1234)).toContain('1')
    expect(percent(42.345)).toBe('42.3%')
    expect(dateTime()).toBe('Never')
    expect(
      relativeTime('2026-01-06T00:00:00Z', new Date('2026-01-11T00:00:00Z')),
    ).toBe('5 days ago')
  })

  it('escapes CSV values safely', () => {
    expect(toCsv(['Name', 'Value'], [['DNS, public', '"trusted"']])).toBe(
      'Name,Value\r\n"DNS, public","""trusted"""',
    )
  })

  it('presents each deployment environment consistently', () => {
    expect(instancePresentation('development')).toMatchObject({
      label: 'Development',
      accent: '#2878c8',
    })
    expect(instancePresentation('staging').label).toBe('Staging')
    expect(instancePresentation('production').accent).toBe('#d84b47')
  })

  it('uses development only for local instances', () => {
    expect(localInstanceEnvironment('localhost')).toBe('development')
    expect(localInstanceEnvironment('127.0.0.1')).toBe('development')
    expect(localInstanceEnvironment('::1')).toBe('development')
    expect(localInstanceEnvironment('firewall.example.com')).toBe('unknown')
  })
})
