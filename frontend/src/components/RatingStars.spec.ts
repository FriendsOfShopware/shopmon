import { describe, it, expect } from 'vitest';
import { mount } from '@vue/test-utils';
import RatingStars from './RatingStars.vue';

describe('RatingStars', () => {
  it('renders nothing when rating is null', () => {
    const wrapper = mount(RatingStars, {
      props: {
        rating: null,
      },
    });
    expect(wrapper.find('.rating-stars').exists()).toBe(false);
  });

  it('renders stars when rating is provided', () => {
    const wrapper = mount(RatingStars, {
      props: {
        rating: 8,
      },
    });
    expect(wrapper.find('.rating-stars').exists()).toBe(true);
  });

  it('displays correct tooltip text', () => {
    const wrapper = mount(RatingStars, {
      props: {
        rating: 8,
      },
    });
    expect(wrapper.find('[data-tooltip="4 from 5"]').exists()).toBe(true);
  });

  it('renders 5 star elements', () => {
    const wrapper = mount(RatingStars, {
      props: {
        rating: 8,
      },
    });
    const stars = wrapper.findAll('.rating-stars > *');
    expect(stars.length).toBe(5);
  });

  it('renders all empty stars for rating of 0', () => {
    const wrapper = mount(RatingStars, {
      props: {
        rating: 0,
      },
    });
    expect(wrapper.find('.rating-stars').exists()).toBe(true);
  });

  it('renders all full stars for rating of 10', () => {
    const wrapper = mount(RatingStars, {
      props: {
        rating: 10,
      },
    });
    expect(wrapper.find('.rating-stars').exists()).toBe(true);
  });

  it('handles odd ratings with half stars', () => {
    const wrapper = mount(RatingStars, {
      props: {
        rating: 7,
      },
    });
    expect(wrapper.find('.rating-stars').exists()).toBe(true);
  });
});
