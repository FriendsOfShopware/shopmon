import { describe, it, expect, beforeEach, vi } from "vitest";
import { useAlert } from "@/composables/useAlert";

describe("useAlert composable", () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  it("should have null alert initially", () => {
    const { alert } = useAlert();
    expect(alert.value).toBeNull();
  });

  it("should set success alert", () => {
    const { alert, success } = useAlert();
    success("Operation completed successfully");

    expect(alert.value).toEqual({
      title: "Action success",
      message: "Operation completed successfully",
      type: "success",
    });
  });

  it("should set info alert", () => {
    const { alert, info } = useAlert();
    info("Here is some information");

    expect(alert.value).toEqual({
      title: "Information",
      message: "Here is some information",
      type: "info",
    });
  });

  it("should set error alert", () => {
    const { alert, error } = useAlert();
    error("Something went wrong");

    expect(alert.value).toEqual({
      title: "Something went wrong",
      message: "Something went wrong",
      type: "error",
    });
  });

  it("should set warning alert", () => {
    const { alert, warning } = useAlert();
    warning("Please be careful");

    expect(alert.value).toEqual({
      title: "Additional information",
      message: "Please be careful",
      type: "warning",
    });
  });

  it("should clear alert when clear is called", () => {
    const { alert, success, clear } = useAlert();
    success("Test message");
    expect(alert.value).not.toBeNull();

    clear();
    expect(alert.value).toBeNull();
  });

  it("should auto-dismiss success alert after 5 seconds", () => {
    const { alert, success } = useAlert();
    success("Test message");

    expect(alert.value).not.toBeNull();
    vi.advanceTimersByTime(5000);
    expect(alert.value).toBeNull();
  });

  it("should auto-dismiss info alert after 5 seconds", () => {
    const { alert, info } = useAlert();
    info("Test message");

    expect(alert.value).not.toBeNull();
    vi.advanceTimersByTime(5000);
    expect(alert.value).toBeNull();
  });

  it("should NOT auto-dismiss error alert", () => {
    const { alert, error } = useAlert();
    error("Test message");

    expect(alert.value).not.toBeNull();
    vi.advanceTimersByTime(5000);
    expect(alert.value).not.toBeNull();
  });

  it("should NOT auto-dismiss warning alert", () => {
    const { alert, warning } = useAlert();
    warning("Test message");

    expect(alert.value).not.toBeNull();
    vi.advanceTimersByTime(5000);
    expect(alert.value).not.toBeNull();
  });

  it("should reset timer when new alert is shown", () => {
    const { alert, success } = useAlert();
    success("First message");

    vi.advanceTimersByTime(3000);
    success("Second message");

    vi.advanceTimersByTime(3000);
    expect(alert.value).not.toBeNull();

    vi.advanceTimersByTime(2000);
    expect(alert.value).toBeNull();
  });
});
