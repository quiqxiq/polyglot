import { useState } from 'react'
import { describe, expect, it, vi } from 'vitest'
import { render } from 'vitest-browser-react'
import { userEvent } from 'vitest/browser'
import { ResourceDatalistInput } from './resource-datalist-input'

// Wrapper terkontrol agar ketikan benar-benar masuk ke input.
function ControlledHarness({
  onChange,
  options,
}: {
  onChange: (v: string) => void
  options?: string[]
}) {
  const [value, setValue] = useState('')
  return (
    <ResourceDatalistInput
      value={value}
      onChange={(v) => {
        setValue(v)
        onChange(v)
      }}
      options={options}
      placeholder='Parent queue'
    />
  )
}

describe('ResourceDatalistInput', () => {
  it('links the input to a datalist containing the router options', async () => {
    const { getByPlaceholder } = await render(
      <ControlledHarness onChange={() => {}} options={['pool-a', 'pool-b']} />
    )

    // Input dengan atribut `list` dipetakan Chromium ke role combobox,
    // jadi query memakai placeholder, bukan role.
    const input = getByPlaceholder('Parent queue')
    await expect.element(input).toBeInTheDocument()
    await expect.element(input).toHaveAttribute('list')

    const listId = input.element().getAttribute('list')
    expect(listId).toBeTruthy()

    const datalist = document.getElementById(listId!)
    expect(datalist).not.toBeNull()
    expect(datalist?.tagName.toLowerCase()).toBe('datalist')

    const values = Array.from(datalist!.querySelectorAll('option')).map(
      (opt) => opt.getAttribute('value'),
    )
    expect(values).toEqual(['pool-a', 'pool-b'])
  })

  it('emits freely typed text through onChange even without a suggestion', async () => {
    const onChange = vi.fn()
    const { getByPlaceholder } = await render(
      <ControlledHarness onChange={onChange} options={['pool-a', 'pool-b']} />
    )

    const input = getByPlaceholder('Parent queue')
    await userEvent.type(input, 'pool-z')

    expect(onChange).toHaveBeenLastCalledWith('pool-z')
    await expect.element(input).toHaveValue('pool-z')
  })

  it('shows the loading placeholder while options are being fetched', async () => {
    const { getByPlaceholder } = await render(
      <ResourceDatalistInput
        onChange={() => {}}
        optionsLoading
        placeholder='Parent queue'
      />
    )

    const loading = getByPlaceholder('Memuat dari router…')
    await expect.element(loading).toBeInTheDocument()
  })

  it('renders as a plain input when no options exist (no device / fetch failed)', async () => {
    const { getByPlaceholder } = await render(
      <ResourceDatalistInput onChange={() => {}} placeholder='Parent queue' />
    )

    const input = getByPlaceholder('Parent queue')
    await expect.element(input).toHaveAttribute('list')

    const datalists = document.querySelectorAll('datalist option')
    expect(datalists.length).toBe(0)
  })
})
