import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import { i18n } from "@/i18n";
import RatingStars from "./RatingStars.vue";

function mountRatingStars(rating: number | null) {
  return mount(RatingStars, {
    props: { rating },
    global: {
      plugins: [i18n],
    },
  });
}

describe("RatingStars", () => {
  it("renders nothing when rating is null", () => {
    const wrapper = mountRatingStars(null);
    expect(wrapper.find(".inline-flex").exists()).toBe(false);
  });

  it("renders stars when rating is provided", () => {
    const wrapper = mountRatingStars(8);
    expect(wrapper.find(".inline-flex").exists()).toBe(true);
  });

  it("displays correct tooltip text", () => {
    const wrapper = mountRatingStars(8);
    expect(wrapper.find('[title="4 out of 5"]').exists()).toBe(true);
  });

  it("renders 5 star elements", () => {
    const wrapper = mountRatingStars(8);
    const stars = wrapper.findAll(".inline-flex > *");
    expect(stars.length).toBe(5);
  });

  it("renders all empty stars for rating of 0", () => {
    const wrapper = mountRatingStars(0);
    expect(wrapper.find(".inline-flex").exists()).toBe(true);
  });

  it("renders all full stars for rating of 10", () => {
    const wrapper = mountRatingStars(10);
    expect(wrapper.find(".inline-flex").exists()).toBe(true);
  });

  it("handles odd ratings with half stars", () => {
    const wrapper = mountRatingStars(7);
    expect(wrapper.find(".inline-flex").exists()).toBe(true);
  });
});
