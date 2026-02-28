import { describe, it, expect } from "vitest"
import { cn } from "./utils"

describe("cn utility", () => {
  it("merges class names", () => {
    const result = cn("foo", "bar")
    expect(result).toBe("foo bar")
  })

  it("handles conditional classes", () => {
    const result = cn("foo", false && "bar", "baz")
    expect(result).toBe("foo baz")
  })

  it("handles empty inputs", () => {
    const result = cn()
    expect(result).toBe("")
  })
})
