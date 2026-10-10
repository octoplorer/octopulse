export function isDeleteConfirmed(input: string, resourceName: string, caseSensitive = true): boolean {
  if (!resourceName)
    return false
  return caseSensitive ? input === resourceName : input.toLocaleLowerCase() === resourceName.toLocaleLowerCase()
}
