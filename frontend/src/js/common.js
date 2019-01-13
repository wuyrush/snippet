const VOIDS = [undefined, null];

function nonNull(obj) {
  return VOIDS.indexOf(obj) == -1;
};

export { nonNull };
