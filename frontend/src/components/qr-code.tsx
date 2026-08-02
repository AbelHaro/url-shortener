import { useRef } from "react";
import { QRCodeCanvas } from "qrcode.react";

import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

type QRCodeProps = {
  value: string;
  size?: number;
  fileName?: string;
  showDownload?: boolean;
  className?: string;
};

export function QRCode({
  value,
  size = 256,
  fileName = "qr-code.png",
  showDownload = true,
  className,
}: QRCodeProps) {
  const canvasRef = useRef<HTMLCanvasElement | null>(null);

  const handleDownload = () => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const link = document.createElement("a");
    link.download = fileName;
    link.href = canvas.toDataURL("image/png");
    link.click();
  };

  return (
    <div className={cn("flex w-full flex-col items-center gap-4", className)}>
      <div className="max-w-full rounded-md bg-white p-2 shadow-sm">
        <QRCodeCanvas
          value={value}
          size={size}
          level="M"
          marginSize={8}
          ref={canvasRef}
          title={`QR code for ${value}`}
          className="h-auto max-w-full"
        />
      </div>
      {showDownload && (
        <Button type="button" variant="secondary" onClick={handleDownload} className="w-full">
          Download QR code
        </Button>
      )}
    </div>
  );
}
