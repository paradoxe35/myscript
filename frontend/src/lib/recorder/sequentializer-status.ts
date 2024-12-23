export default class SequentializerStatus {
  private pendingSequentializer: boolean = false;
  private eventTarget: EventTarget;
  private eventName = "disable-pending-sequentializer";

  constructor() {
    this.eventTarget = new EventTarget();
  }

  enableInPending() {
    this.pendingSequentializer = true;
  }

  disableInPending() {
    this.pendingSequentializer = false;
    this.eventTarget.dispatchEvent(new Event(this.eventName));
  }

  async canExport() {
    return new Promise<any>((resolve) => {
      if (this.pendingSequentializer === false) {
        resolve(true);
      } else {
        this.eventTarget.addEventListener(this.eventName, resolve, {
          once: true,
        });
      }
    });
  }
}
