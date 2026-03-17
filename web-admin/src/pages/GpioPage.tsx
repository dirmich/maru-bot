import { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Button } from '@/components/ui/button';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Cpu, Save, Plus, Trash, RefreshCw } from 'lucide-react';
import { toast } from 'sonner';
import { useTranslation } from "@/lib/i18n";
import { Switch } from "@/components/ui/switch";

import { GpioSchematic, pinData } from '@/components/gpio-schematic';
import { ConfirmDialog } from '@/components/ui-custom-dialog';

interface PinConfig {
    pin: number;
    mode: string;
    label: string;
    level?: number;
}

export function GpioPage() {
    const t = useTranslation();
    const [configuredPins, setConfiguredPins] = useState<PinConfig[]>([]);
    const [selectedPin, setSelectedPin] = useState<number | undefined>(undefined);
    const [confirmOpen, setConfirmOpen] = useState(false);
    const [pendingPin, setPendingPin] = useState<number | null>(null);

    const handleAddPin = () => {
        // Find next available pin or just use 0 as placeholder
        const newPin = { pin: 0, mode: 'OUT', label: 'New Device' };
        setConfiguredPins([...configuredPins, newPin]);
        setSelectedPin(undefined); // Reset selection to show all
    };

    const handlePinClick = (pin: number) => {
        const pinInfo = pinData.find(p => p.number === pin);
        if (pinInfo && (pinInfo.type === 'power' || pinInfo.type === 'ground')) {
            toast.error(t.gpio_cannot_configure);
            return;
        }

        const existing = configuredPins.find(p => p.pin === pin);
        if (existing) {
            setSelectedPin(pin);
        } else {
            // Ask to add using ConfirmDialog
            setPendingPin(pin);
            setConfirmOpen(true);
        }
    };

    const handleConfirmAdd = () => {
        if (pendingPin !== null) {
            const newPin = { pin: pendingPin, mode: 'OUT', label: `GPIO ${pendingPin}` };
            setConfiguredPins([...configuredPins, newPin]);
            setSelectedPin(pendingPin);
            setPendingPin(null);
        }
    };

    const handleRemovePin = (index: number) => {
        const newPins = [...configuredPins];
        newPins.splice(index, 1);
        setConfiguredPins(newPins);
    };

    const handleUpdatePin = (index: number, field: keyof PinConfig, value: any) => {
        const newPins = [...configuredPins];
        (newPins[index] as any)[field] = value;
        setConfiguredPins(newPins);
    }

    useEffect(() => {
        fetchGpio();
    }, []);

    const fetchGpio = async () => {
        try {
            const res = await fetch('/api/gpio');
            if (res.ok) {
                const data = await res.json();
                // Data is now map[string]int (flattened)
                const pins: PinConfig[] = Object.entries(data).map(([label, pin]: [string, any]) => {
                    return {
                        pin: pin as number,
                        mode: isInput(label) ? 'IN' : 'OUT',
                        label
                    };
                });
                if (pins.length >= 0) setConfiguredPins(pins);
            }
        } catch (e) {
            console.error("Failed to fetch GPIO", e);
        }
    };

    const isInput = (label: string) => {
        const l = label.toLowerCase();
        return l === 'button' || l === 'sensor' || l.startsWith('button_') || l.startsWith('sensor_') || l.endsWith('_button') || l.endsWith('_sensor');
    };

    const handleToggle = async (pin: number, index: number) => {
        try {
            const res = await fetch('/api/gpio/toggle', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ pin }),
            });
            if (res.ok) {
                const data = await res.json();
                const newPins = [...configuredPins];
                newPins[index].level = data.level;
                setConfiguredPins(newPins);
                
                const verb = data.action === 'read' ? 'is' : 'toggled to';
                toast.success(`Pin ${pin} ${verb} ${data.level === 1 ? 'HIGH' : 'LOW'}`);
            }
        } catch (e) {
            toast.error("Toggle failed");
        }
    };

    const handleSave = async () => {
        try {
            // Transform PinConfig[] back to flat map for backend
            const pinMap: Record<string, number> = {};
            configuredPins.forEach(p => {
                pinMap[p.label] = p.pin;
            });

            const res = await fetch('/api/gpio', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(pinMap),
            });

            if (res.ok) {
                toast.success(t.gpio_save_success);
                setSelectedPin(undefined); // Return to view all pins
            } else {
                toast.error('Error (HTTP ' + res.status + ')');
            }
        } catch (e) {
            toast.error('Network Error');
        }
    };

    return (
        <div className="p-6 max-w-6xl mx-auto space-y-6">
            <header className="mb-6">
                <h1 className="text-2xl font-bold flex items-center gap-2">
                    <Cpu className="text-orange-600" /> {t.gpio_title}
                </h1>
                <p className="text-sm text-slate-500">{t.gpio_desc}</p>
            </header>

            <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
                <div className="space-y-6">
                    <Card className="border-none shadow-lg">
                        <CardHeader>
                            <CardTitle className="text-lg">{t.gpio_schematic}</CardTitle>
                            <CardDescription>{t.gpio_schematic_desc}</CardDescription>
                        </CardHeader>
                        <CardContent>
                            <GpioSchematic
                                configuredPins={configuredPins.map(p => p.pin)}
                                selectedPin={selectedPin}
                                onPinClick={handlePinClick}
                            />
                        </CardContent>
                    </Card>
                </div>

                <div className="space-y-6 flex flex-col">
                    <Card className="border-none shadow-lg flex-1">
                        <CardHeader className="flex flex-row items-center justify-between">
                            <div>
                                <CardTitle className="text-lg">{t.gpio_configured_devices}</CardTitle>
                                <CardDescription>{t.gpio_configured_desc}</CardDescription>
                            </div>
                            <div className="flex gap-2">
                                {selectedPin !== undefined && (
                                    <Button size="sm" variant="outline" onClick={() => setSelectedPin(undefined)}>
                                        {t.gpio_view_all}
                                    </Button>
                                )}
                                <Button size="sm" onClick={handleAddPin} className="bg-orange-600 hover:bg-orange-700 text-white">
                                    <Plus className="w-4 h-4 mr-1" /> {t.gpio_add}
                                </Button>
                            </div>
                        </CardHeader>
                        <CardContent>
                            <Table>
                                <TableHeader>
                                    <TableRow>
                                        <TableHead className="w-12"></TableHead>
                                        <TableHead className="w-24">{t.gpio_pin}</TableHead>
                                        <TableHead className="w-28">{t.gpio_mode}</TableHead>
                                        <TableHead>{t.gpio_label}</TableHead>
                                        <TableHead className="w-10"></TableHead>
                                    </TableRow>
                                </TableHeader>
                                <TableBody>
                                    {configuredPins.map((item, idx) => {
                                        if (selectedPin !== undefined && item.pin !== selectedPin) return null;
                                        return (
                                            <TableRow key={idx}>
                                                <TableCell className="px-2">
                                                    {item.mode === 'OUT' ? (
                                                        <Switch
                                                            checked={item.level === 1}
                                                            onCheckedChange={() => handleToggle(item.pin, idx)}
                                                        />
                                                    ) : item.mode === 'IN' ? (
                                                        <div className="flex items-center gap-2">
                                                            <span className={`w-3 h-3 rounded-full ${item.level === 1 ? 'bg-green-500 shadow-[0_0_8px_rgba(34,197,94,0.6)]' : 'bg-slate-300'}`} />
                                                            <Button 
                                                                variant="ghost" 
                                                                size="icon" 
                                                                className="h-7 w-7 text-slate-400 hover:text-blue-500"
                                                                onClick={() => handleToggle(item.pin, idx)}
                                                            >
                                                                <RefreshCw className="w-3.5 h-3.5" />
                                                            </Button>
                                                        </div>
                                                    ) : null}
                                                </TableCell>
                                                <TableCell className="font-mono">
                                                    <Select
                                                        value={item.pin.toString()}
                                                        onValueChange={(v) => handleUpdatePin(idx, 'pin', parseInt(v))}
                                                    >
                                                        <SelectTrigger className="w-20 h-8 text-xs font-mono">
                                                            <SelectValue />
                                                        </SelectTrigger>
                                                        <SelectContent>
                                                            {pinData
                                                                .filter(p => p.type !== 'power' && p.type !== 'ground')
                                                                .sort((a, b) => a.number - b.number)
                                                                .map(p => {
                                                                    const isUsedByOthers = configuredPins.some((cp, cpidx) => cp.pin === p.number && cpidx !== idx);
                                                                    return (
                                                                        <SelectItem
                                                                            key={p.number}
                                                                            value={p.number.toString()}
                                                                            disabled={isUsedByOthers}
                                                                        >
                                                                            {p.number} {isUsedByOthers ? '(Used)' : ''}
                                                                        </SelectItem>
                                                                    );
                                                                })}
                                                        </SelectContent>
                                                    </Select>
                                                </TableCell>
                                                <TableCell>
                                                    <Select
                                                        value={item.mode}
                                                        onValueChange={(v) => handleUpdatePin(idx, 'mode', v)}
                                                    >
                                                        <SelectTrigger className="w-24 h-8 text-xs">
                                                            <SelectValue />
                                                        </SelectTrigger>
                                                        <SelectContent>
                                                            <SelectItem value="OUT">OUT</SelectItem>
                                                            <SelectItem value="IN">IN</SelectItem>
                                                            <SelectItem value="PWM">PWM</SelectItem>
                                                            <SelectItem value="I2C">I2C</SelectItem>
                                                            <SelectItem value="SPI">SPI</SelectItem>
                                                        </SelectContent>
                                                    </Select>
                                                </TableCell>
                                                <TableCell>
                                                    <input
                                                        className="bg-transparent border-b border-dashed border-slate-300 dark:border-slate-700 focus:outline-none focus:border-orange-500 w-full text-sm"
                                                        value={item.label}
                                                        onChange={(e) => handleUpdatePin(idx, 'label', e.target.value)}
                                                    />
                                                </TableCell>
                                                <TableCell>
                                                    <Button
                                                        variant="ghost"
                                                        size="icon"
                                                        className="h-8 w-8 text-slate-400 hover:text-red-500"
                                                        onClick={() => handleRemovePin(idx)}
                                                    >
                                                        <Trash className="w-4 h-4" />
                                                    </Button>
                                                </TableCell>
                                            </TableRow>
                                        );
                                    })}
                                    {configuredPins.length === 0 && (
                                        <TableRow>
                                            <TableCell colSpan={4} className="text-center text-slate-400 py-8">
                                                {t.gpio_no_pins}
                                            </TableCell>
                                        </TableRow>
                                    )}
                                </TableBody>
                            </Table>
                        </CardContent>
                        <CardFooter className="justify-end border-t pt-4">
                            <Button onClick={handleSave} className="bg-blue-600 hover:bg-blue-700 text-white">
                                <Save className="w-4 h-4 mr-2" /> {t.gpio_save}
                            </Button>
                        </CardFooter>
                    </Card>
                </div>
            </div>

            <ConfirmDialog
                open={confirmOpen}
                onOpenChange={setConfirmOpen}
                title={t.gpio_add_confirm_title}
                description={`${t.gpio_add_confirm_desc} ${pendingPin}${t.gpio_add_confirm_suffix || "?"}`}
                onConfirm={handleConfirmAdd}
            />
        </div>
    );
}
