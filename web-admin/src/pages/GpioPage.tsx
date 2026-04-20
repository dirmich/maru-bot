import { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { Table, TableBody, TableCell, TableRow } from '@/components/ui/table';
import { Button } from '@/components/ui/button';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Cpu, Save, Plus, Trash, RefreshCw } from 'lucide-react';
import { toast } from 'sonner';
import { useTranslation } from "@/lib/i18n";
import { Switch } from "@/components/ui/switch";
import { authenticatedFetch } from "@/lib/auth";

import { GpioSchematic, pinData } from '@/components/gpio-schematic';
import { ConfirmDialog } from '@/components/ui-custom-dialog';

interface PinConfig {
    pin: number;
    mode: string;
    label: string;
    group: string;
    level?: number;
}

export function GpioPage() {
    const t = useTranslation();
    const [configuredPins, setConfiguredPins] = useState<PinConfig[]>([]);
    const [selectedPin, setSelectedPin] = useState<number | undefined>(undefined);
    const [confirmOpen, setConfirmOpen] = useState(false);
    const [pendingPin, setPendingPin] = useState<number | null>(null);

    const handleAddPin = (group: string = "") => {
        const newPin = { pin: 0, mode: 'OUT', label: 'New Device', group };
        setConfiguredPins([...configuredPins, newPin]);
        setSelectedPin(undefined);
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
            setPendingPin(pin);
            setConfirmOpen(true);
        }
    };

    const handleConfirmAdd = () => {
        if (pendingPin !== null) {
            const newPin = { pin: pendingPin, mode: 'OUT', label: `GPIO ${pendingPin}`, group: "" };
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
            const res = await authenticatedFetch('/api/gpio');
            if (res.ok) {
                const data = await res.json();
                const flattened: PinConfig[] = [];
                
                const parseRecursive = (obj: any, group: string = "") => {
                    for (const key in obj) {
                        const val = obj[key];
                        if (typeof val === 'object' && val !== null) {
                            parseRecursive(val, group ? `${group}_${key}` : key);
                        } else if (typeof val === 'number') {
                            flattened.push({
                                pin: val,
                                label: key,
                                group: group,
                                mode: isInput(key) || isInput(group) ? 'IN' : 'OUT'
                            });
                        }
                    }
                };

                parseRecursive(data);
                setConfiguredPins(flattened);
            }
        } catch (e) {
            console.error("Failed to fetch GPIO", e);
        }
    };

    const isInput = (label: string) => {
        const l = label.toLowerCase();
        return l === 'button' || l === 'sensor' || l.startsWith('button_') || l.startsWith('sensor_') || l.endsWith('_button') || l.endsWith('_sensor');
    };

    const handleToggle = async (pin: number, label: string, group: string) => {
        try {
            const res = await authenticatedFetch('/api/gpio/toggle', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ pin }),
            });
            if (res.ok) {
                const data = await res.json();
                setConfiguredPins(prev => prev.map(p => 
                    (p.pin === pin && p.label === label && p.group === group) ? { ...p, level: data.level } : p
                ));
                
                const verb = data.action === 'read' ? 'is' : 'toggled to';
                toast.success(`Pin ${pin} ${verb} ${data.level === 1 ? 'HIGH' : 'LOW'}`);
            }
        } catch (e) {
            toast.error("Toggle failed");
        }
    };

    const handleAddGroup = () => {
        const newGroupName = `Group_${Object.keys(groups).length + 1}`;
        // To create a group, we need at least one pin or just a placeholder in UI
        handleAddPin(newGroupName);
    };

    const handleRenameGroup = (oldName: string, newName: string) => {
        if (!newName || oldName === newName) return;
        setConfiguredPins(prev => prev.map(p => 
            p.group === oldName ? { ...p, group: newName } : p
        ));
    };

    const handleRemoveGroup = (groupName: string) => {
        setConfiguredPins(prev => prev.filter(p => p.group !== groupName));
    };

    const isInvalidPin = (pin: number) => {
        const pinInfo = pinData.find(p => p.number === pin);
        return pinInfo && (pinInfo.type === 'power' || pinInfo.type === 'ground');
    };

    const handleSave = async () => {
        const invalidPins = configuredPins.filter(p => isInvalidPin(p.pin));
        if (invalidPins.length > 0) {
            toast.error("VCC/GND 핀은 사용할 수 없습니다. 설정을 수정해 주세요.");
            return;
        }

        try {
            const pinMap: Record<string, number> = {};
            configuredPins.forEach(p => {
                const fullKey = p.group ? `${p.group}_${p.label}` : p.label;
                pinMap[fullKey] = p.pin;
            });

            const res = await authenticatedFetch('/api/gpio', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(pinMap),
            });

            if (res.ok) {
                toast.success(t.gpio_save_success);
                setSelectedPin(undefined);
            } else {
                toast.error('Error (HTTP ' + res.status + ')');
            }
        } catch (e) {
            toast.error('Network Error');
        }
    };

    // Group pins by group name
    const groups = configuredPins.reduce((acc, pin) => {
        const groupName = pin.group || "Default";
        if (!acc[groupName]) acc[groupName] = [];
        acc[groupName].push(pin);
        return acc;
    }, {} as Record<string, PinConfig[]>);

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
                        <CardFooter className="justify-end border-t pt-4">
                            <Button 
                                onClick={handleSave} 
                                className="bg-blue-600 hover:bg-blue-700 text-white"
                                disabled={configuredPins.some(p => isInvalidPin(p.pin))}
                            >
                                <Save className="w-4 h-4 mr-2" /> {t.gpio_save}
                            </Button>
                        </CardFooter>
                    </Card>
                </div>

                <div className="space-y-6 flex flex-col">
                    <div className="flex justify-between items-center mb-2">
                        <h2 className="text-xl font-bold text-slate-700 dark:text-slate-300">Configured Devices</h2>
                        <div className="flex gap-2">
                            <Button size="sm" variant="outline" onClick={handleAddGroup} className="text-blue-600 border-blue-200">
                                <Plus className="w-4 h-4 mr-1" /> Add Group
                            </Button>
                            <Button size="sm" onClick={() => handleAddPin("")} className="bg-orange-600 hover:bg-orange-700 text-white">
                                <Plus className="w-4 h-4 mr-1" /> {t.gpio_add}
                            </Button>
                        </div>
                    </div>

                    {Object.entries(groups).map(([groupName, pins]) => (
                        <Card key={groupName} className="border-none shadow-md bg-slate-50/50 dark:bg-slate-900/50 border border-slate-200 dark:border-slate-800">
                            <CardHeader className="py-3 px-4 flex flex-row items-center justify-between border-b bg-slate-100/50 dark:bg-slate-800/50 rounded-t-lg">
                                <div className="flex items-center gap-2 flex-1">
                                    {groupName === "Default" ? (
                                        <CardTitle className="text-sm font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wider">
                                            General
                                        </CardTitle>
                                    ) : (
                                        <input
                                            className="text-sm font-semibold text-blue-600 dark:text-blue-400 uppercase tracking-wider bg-transparent border-b border-transparent focus:border-blue-500 focus:outline-none w-1/2"
                                            value={groupName}
                                            onChange={(e) => handleRenameGroup(groupName, e.target.value)}
                                        />
                                    )}
                                    {groupName !== "Default" && (
                                        <span className="text-[10px] bg-blue-100 text-blue-600 dark:bg-blue-900/30 dark:text-blue-400 px-2 py-0.5 rounded-full font-bold">GROUP</span>
                                    )}
                                </div>
                                <div className="flex items-center gap-1">
                                    <Button 
                                        variant="ghost" 
                                        size="icon" 
                                        className="h-7 w-7 text-slate-400 hover:text-orange-500"
                                        onClick={() => handleAddPin(groupName === "Default" ? "" : groupName)}
                                        title="Add Pin to Group"
                                    >
                                        <Plus className="w-4 h-4" />
                                    </Button>
                                    {groupName !== "Default" && (
                                        <Button 
                                            variant="ghost" 
                                            size="icon" 
                                            className="h-7 w-7 text-slate-400 hover:text-red-500"
                                            onClick={() => handleRemoveGroup(groupName)}
                                        >
                                            <Trash className="w-3.5 h-3.5" />
                                        </Button>
                                    )}
                                </div>
                            </CardHeader>
                            <CardContent className="p-0">
                                <Table>
                                    <TableBody>
                                        {pins.map((item, localIdx) => {
                                            const globalIdx = configuredPins.findIndex(p => p === item);
                                            const isError = isInvalidPin(item.pin);
                                            if (selectedPin !== undefined && item.pin !== selectedPin) return null;
                                            return (
                                                <TableRow key={`${groupName}-${localIdx}`} className={`border-b last:border-0 transition-colors ${isError ? 'bg-red-50 dark:bg-red-900/20 hover:bg-red-100 dark:hover:bg-red-900/30' : 'hover:bg-white dark:hover:bg-slate-800'}`}>
                                                    <TableCell className="w-12 px-3 py-2 text-center">
                                                        {item.mode === 'OUT' ? (
                                                            <Switch
                                                                checked={item.level === 1}
                                                                onCheckedChange={() => handleToggle(item.pin, item.label, item.group)}
                                                            />
                                                        ) : (
                                                            <div className="flex items-center justify-center gap-2">
                                                                <span className={`w-3 h-3 rounded-full ${item.level === 1 ? 'bg-green-500 shadow-[0_0_8px_rgba(34,197,94,0.6)]' : 'bg-slate-300'}`} />
                                                                <Button 
                                                                    variant="ghost" 
                                                                    size="icon" 
                                                                    className="h-7 w-7 text-slate-400 hover:text-blue-500"
                                                                    onClick={() => handleToggle(item.pin, item.label, item.group)}
                                                                >
                                                                    <RefreshCw className="w-3.5 h-3.5" />
                                                                </Button>
                                                            </div>
                                                        )}
                                                    </TableCell>
                                                    <TableCell className="w-20 px-3 py-2">
                                                        <Select
                                                            value={item.pin.toString()}
                                                            onValueChange={(v) => handleUpdatePin(globalIdx, 'pin', parseInt(v))}
                                                        >
                                                            <SelectTrigger className={`h-8 text-xs font-mono border-none shadow-none bg-transparent hover:bg-slate-100 dark:hover:bg-slate-800 ${isError ? 'text-red-500 font-bold' : ''}`}>
                                                                <SelectValue />
                                                            </SelectTrigger>
                                                            <SelectContent>
                                                                {pinData
                                                                    // .filter(p => p.type !== 'power' && p.type !== 'ground') // Let them see it but mark error if they pick it, or just keep filtering. 
                                                                    // Actually, the user asked to show error IF configured pin is VCC/GND. 
                                                                    // If we filter them out from the select, they can't pick them anyway.
                                                                    // But they might be there from a config migration Or the user might want to know WHY they can't pick them.
                                                                    .sort((a, b) => a.number - b.number)
                                                                    .map(p => (
                                                                        <SelectItem 
                                                                            key={p.number} 
                                                                            value={p.number.toString()}
                                                                            className={p.type === 'power' || p.type === 'ground' ? 'text-red-400 italic' : ''}
                                                                        >
                                                                            {p.number} {p.type === 'power' ? '(VCC)' : p.type === 'ground' ? '(GND)' : ''}
                                                                        </SelectItem>
                                                                    ))}
                                                            </SelectContent>
                                                        </Select>
                                                        {isError && <div className="text-[9px] text-red-500 font-bold mt-0.5 ml-1">VCC/GND Error</div>}
                                                    </TableCell>
                                                    <TableCell className="px-3 py-2">
                                                        <input
                                                            className="bg-transparent border-none focus:ring-0 w-full text-sm font-medium"
                                                            value={item.label}
                                                            placeholder="Label"
                                                            onChange={(e) => handleUpdatePin(globalIdx, 'label', e.target.value)}
                                                        />
                                                        <div className="flex gap-2 mt-1">
                                                            <input
                                                                className="bg-transparent border-none focus:ring-0 text-[10px] text-slate-400 w-1/2"
                                                                value={item.group}
                                                                placeholder="Group"
                                                                onChange={(e) => handleUpdatePin(globalIdx, 'group', e.target.value)}
                                                            />
                                                            <Select
                                                                value={item.mode}
                                                                onValueChange={(v) => handleUpdatePin(globalIdx, 'mode', v)}
                                                            >
                                                                <SelectTrigger className="h-4 p-0 border-none shadow-none text-[10px] text-blue-500 font-bold w-12 bg-transparent">
                                                                    <SelectValue />
                                                                </SelectTrigger>
                                                                <SelectContent>
                                                                    <SelectItem value="OUT">OUT</SelectItem>
                                                                    <SelectItem value="IN">IN</SelectItem>
                                                                </SelectContent>
                                                            </Select>
                                                        </div>
                                                    </TableCell>
                                                    <TableCell className="w-10 px-3 py-2">
                                                        <Button
                                                            variant="ghost"
                                                            size="icon"
                                                            className="h-8 w-8 text-slate-300 hover:text-red-500"
                                                            onClick={() => handleRemovePin(globalIdx)}
                                                        >
                                                            <Trash className="w-4 h-4" />
                                                        </Button>
                                                    </TableCell>
                                                </TableRow>
                                            );
                                        })}
                                    </TableBody>
                                </Table>
                            </CardContent>
                        </Card>
                    ))}

                    {configuredPins.length === 0 && (
                        <div className="text-center text-slate-400 py-12 bg-slate-50 dark:bg-slate-900 rounded-xl border-2 border-dashed border-slate-200 dark:border-slate-800">
                            {t.gpio_no_pins}
                        </div>
                    )}
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
