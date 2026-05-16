let _push = null;
export const registerToast = (fn) => { _push = fn; };
export const showToast = (msg, type = 'info') => { _push?.(msg, type); };
