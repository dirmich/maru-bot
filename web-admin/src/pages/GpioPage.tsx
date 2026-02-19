import { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Button } from '@/components/ui/button';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Cpu, Save, Plus, Trash } from 'lucide-react';
import { toast } from 'sonner';
import { useTranslation } from "@/lib/i18n";

// Placeholder for GpioSchematic
const GpioSchematic = ({ configuredPins, t }: { configuredPins: number[], t: any }) => (
    <div className="bg-slate-100 p-8 rounded-lg text-center border-2 border-dashed border-slate-300">
        <h3 className="font-semibold text-slate-500 mb-2">{t.gpio_schematic}</h3>
        <p className="text-xs text-slate-400">{t.gpio_schematic_desc}</p>
        <div className="mt-4 grid grid-cols-2 gap-2 max-w-xs mx-auto text-xs font-mono">
            {[...Array(40)].map((_, i) => (
                <div key={i} className={`h-4 w-4 rounded-full mx-auto ${configuredPins.includes(i + 1) ? 'bg-orange-500' : 'bg-slate-300'}`}></div>
            ))}
        </div>
    </div>
);

interface PinConfig {
    pin: number;
    mode: string;
    label: string;
}

export function GpioPage() {
    const t = useTranslation();
    const [configuredPins, setConfiguredPins] = useState<PinConfig[]>([]);

    const handleAddPin = () => {
        // Find next available pin or just use 0 as placeholder
        const newPin = { pin: 0, mode: 'OUT', label: 'New Device' };
        setConfiguredPins([...configuredPins, newPin]);
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
                // If data is in the format expected by the frontend
                // Backend returns map[string]interface{}
                // We might need to transform it if the frontend expects PinConfig[]
                const pins: PinConfig[] = Object.entries(data).map(([label, pin]: [string, any]) => {
                    if (typeof pin === 'number') {
                        return { pin, mode: 'OUT', label };
                    }
                    return { pin: 0, mode: 'OUT', label }; // Fallback
                });
                if (pins.length > 0) setConfiguredPins(pins);
            }
        } catch (e) {
            console.error("Failed to fetch GPIO", e);
        }
    };

    const handleSave = async () => {
        try {
            // Transform PinConfig[] back to map[string]interface{} for backend
            const pinMap: Record<string, any> = {};
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
                            <GpioSchematic configuredPins={configuredPins.map(p => p.pin)} t={t} />
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
                            <Button size="sm" onClick={handleAddPin} className="bg-orange-600 hover:bg-orange-700 text-white">
                                <Plus className="w-4 h-4 mr-1" /> {t.gpio_add}
                            </Button>
                        </CardHeader>
                        <CardContent>
                            <Table>
                                <TableHeader>
                                    <TableRow>
                                        <TableHead>{t.gpio_pin}</TableHead>
                                        <TableHead>{t.gpio_mode}</TableHead>
                                        <TableHead>{t.gpio_label}</TableHead>
                                        <TableHead className="w-10"></TableHead>
                                    </TableRow>
                                </TableHeader>
                                <TableBody>
                                    {configuredPins.map((item, idx) => (
                                        <TableRow key={idx}>
                                            <TableCell className="font-mono">
                                                <input
                                                    type="number"
                                                    className="w-12 bg-transparent border-b border-dashed text-center"
                                                    value={item.pin}
                                                    onChange={(e) => handleUpdatePin(idx, 'pin', parseInt(e.target.value))}
                                                />
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
                                    ))}
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
        </div>
    );
}
