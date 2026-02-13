'use client';

import { Card } from "@/components/ui/card";
import { cn } from "@/lib/utils";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";

interface PinProps {
    number: number;
    label: string;
    type: 'power' | 'ground' | 'gpio' | 'special';
    name?: string;
    isConfigured?: boolean;
}

const pinData: PinProps[] = [
    // Left Side (Odd)
    { number: 1, label: "3.3V", type: 'power' },
    { number: 3, label: "GPIO 2 (SDA)", type: 'gpio', name: "I2C SDA" },
    { number: 5, label: "GPIO 3 (SCL)", type: 'gpio', name: "I2C SCL" },
    { number: 7, label: "GPIO 4", type: 'gpio' },
    { number: 9, label: "Ground", type: 'ground' },
    { number: 11, label: "GPIO 17", type: 'gpio' },
    { number: 13, label: "GPIO 27", type: 'gpio' },
    { number: 15, label: "GPIO 22", type: 'gpio' },
    { number: 17, label: "3.3V", type: 'power' },
    { number: 19, label: "GPIO 10", type: 'gpio' },
    { number: 21, label: "GPIO 9", type: 'gpio' },
    { number: 23, label: "GPIO 11", type: 'gpio' },
    { number: 25, label: "Ground", type: 'ground' },
    { number: 27, label: "ID_SD", type: 'special' },
    { number: 29, label: "GPIO 5", type: 'gpio' },
    { number: 31, label: "GPIO 6", type: 'gpio' },
    { number: 33, label: "GPIO 13", type: 'gpio' },
    { number: 35, label: "GPIO 19", type: 'gpio' },
    { number: 37, label: "GPIO 26", type: 'gpio' },
    { number: 39, label: "Ground", type: 'ground' },

    // Right Side (Even)
    { number: 2, label: "5V", type: 'power' },
    { number: 4, label: "5V", type: 'power' },
    { number: 6, label: "Ground", type: 'ground' },
    { number: 8, label: "GPIO 14", type: 'gpio' },
    { number: 10, label: "GPIO 15", type: 'gpio' },
    { number: 12, label: "GPIO 18", type: 'gpio' },
    { number: 14, label: "Ground", type: 'ground' },
    { number: 16, label: "GPIO 23", type: 'gpio' },
    { number: 18, label: "GPIO 24", type: 'gpio' },
    { number: 20, label: "Ground", type: 'ground' },
    { number: 22, label: "GPIO 25", type: 'gpio' },
    { number: 24, label: "GPIO 8", type: 'gpio' },
    { number: 26, label: "GPIO 7", type: 'gpio' },
    { number: 28, label: "ID_SC", type: 'special' },
    { number: 30, label: "Ground", type: 'ground' },
    { number: 32, label: "GPIO 12", type: 'gpio' },
    { number: 34, label: "Ground", type: 'ground' },
    { number: 36, label: "GPIO 16", type: 'gpio' },
    { number: 38, label: "GPIO 20", type: 'gpio' },
    { number: 40, label: "GPIO 21", type: 'gpio' },
];

export function GpioSchematic({ configuredPins }: { configuredPins: number[] }) {
    const leftPins = pinData.filter(p => p.number % 2 !== 0).sort((a, b) => a.number - b.number);
    const rightPins = pinData.filter(p => p.number % 2 === 0).sort((a, b) => a.number - b.number);

    return (
        <Card className="p-8 bg-slate-900 border-slate-800 shadow-2xl overflow-hidden relative">
            <div className="absolute top-0 right-0 p-4 text-[10px] text-slate-500 font-mono italic">
                Raspberry Pi 4/5 Pinout
            </div>

            <div className="flex justify-center gap-1">
                {/* Left Column Labels */}
                <div className="flex flex-col gap-2 text-right justify-center">
                    {leftPins.map(pin => (
                        <div key={pin.number} className="h-6 flex items-center justify-end pr-2">
                            <span className="text-[10px] font-mono text-slate-400">{pin.label}</span>
                        </div>
                    ))}
                </div>

                {/* Pins Visual */}
                <div className="bg-slate-800 p-3 rounded-md border border-slate-700 flex gap-4 shadow-inner">
                    {/* Left Pins */}
                    <div className="flex flex-col gap-2">
                        {leftPins.map(pin => (
                            <PinItem key={pin.number} pin={pin} isActive={configuredPins.includes(pin.number)} />
                        ))}
                    </div>
                    {/* Right Pins */}
                    <div className="flex flex-col gap-2">
                        {rightPins.map(pin => (
                            <PinItem key={pin.number} pin={pin} isActive={configuredPins.includes(pin.number)} />
                        ))}
                    </div>
                </div>

                {/* Right Column Labels */}
                <div className="flex flex-col gap-2 text-left justify-center">
                    {rightPins.map(pin => (
                        <div key={pin.number} className="h-6 flex items-center justify-start pl-2">
                            <span className="text-[10px] font-mono text-slate-400">{pin.label}</span>
                        </div>
                    ))}
                </div>
            </div>

            <div className="mt-8 grid grid-cols-4 gap-2">
                <LegendItem type="power" label="Power (3.3V/5V)" />
                <LegendItem type="ground" label="Ground" />
                <LegendItem type="gpio" label="Digital GPIO" />
                <LegendItem type="special" label="Special / I2C" />
            </div>
        </Card>
    );
}

function PinItem({ pin, isActive }: { pin: PinProps, isActive: boolean }) {
    const colors = {
        power: "bg-red-500 hover:bg-red-400 shadow-red-900/50",
        ground: "bg-black hover:bg-slate-900 shadow-black/50 border border-slate-700",
        gpio: "bg-orange-500 hover:bg-orange-400 shadow-orange-900/50",
        special: "bg-blue-500 hover:bg-blue-400 shadow-blue-900/50",
    };

    return (
        <TooltipProvider>
            <Tooltip>
                <TooltipTrigger asChild>
                    <div className={cn(
                        "w-6 h-6 rounded-sm cursor-help transition-all transform hover:scale-110 shadow-sm flex items-center justify-center text-[8px] font-bold text-white",
                        colors[pin.type],
                        isActive ? "ring-2 ring-white ring-offset-2 ring-offset-slate-800 scale-105" : ""
                    )}>
                        {pin.number}
                    </div>
                </TooltipTrigger>
                <TooltipContent side="right">
                    <p className="font-bold">{pin.label}</p>
                    <p className="text-xs text-slate-400">Pin {pin.number}</p>
                    {isActive && <p className="text-xs text-emerald-400 mt-1 font-semibold">● Configured</p>}
                </TooltipContent>
            </Tooltip>
        </TooltipProvider>
    );
}

function LegendItem({ type, label }: { type: PinProps['type'], label: string }) {
    const colors = {
        power: "bg-red-500",
        ground: "bg-black border border-slate-700",
        gpio: "bg-orange-500",
        special: "bg-blue-500",
    };
    return (
        <div className="flex items-center gap-2">
            <div className={cn("w-3 h-3 rounded-sm", colors[type])}></div>
            <span className="text-xs text-slate-400">{label}</span>
        </div>
    );
}
