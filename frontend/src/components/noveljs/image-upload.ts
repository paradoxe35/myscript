import { wait } from "@/lib/utils";
import { createImageUpload } from "novel/plugins";
import { toast } from "sonner";

const onUpload = (file: File) => {
  return new Promise(async (resolve) => {
    const blobUrl = URL.createObjectURL(file);
    await wait(100);
    resolve(blobUrl);

    await wait(1000);

    URL.revokeObjectURL(blobUrl);
  });
};

export const uploadFn = createImageUpload({
  onUpload,
  validateFn: (file) => {
    if (!file.type.includes("image/")) {
      toast.error("File type not supported.");
      return false;
    }
    if (file.size / 1024 / 1024 > 20) {
      toast.error("File size too big (max 20MB).");
      return false;
    }
    return true;
  },
});
