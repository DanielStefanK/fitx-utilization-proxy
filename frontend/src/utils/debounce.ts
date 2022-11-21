const debounce = (func: (...args: any[]) => any, timeout: number) => {
  let timer: number;
  return (...args: any[]) => {
    clearTimeout(timer)
    timer = setTimeout(() => {
      func.apply(this, args)
    }, timeout)
  }
}

export default debounce