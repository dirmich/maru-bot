'use client';

import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { create } from "zustand";

interface AlertDialogStore {
    isOpen: boolean;
    title: string;
    description: string;
    onConfirm: () => void;
    show: (title: string, description: string, onConfirm?: () => void) => void;
    hide: () => void;
}

export const useAlertDialog = create<AlertDialogStore>((set) => ({
    isOpen: false,
    title: "",
    description: "",
    onConfirm: () => { },
    show: (title, description, onConfirm) =>
        set({ isOpen: true, title, description, onConfirm: onConfirm || (() => { }) }),
    hide: () => set({ isOpen: false }),
}));

export function AlertDialog() {
    const { isOpen, title, description, hide } = useAlertDialog();

    return (
        <Dialog open={isOpen} onOpenChange={hide}>
            <DialogContent>
                <DialogHeader>
                    <DialogTitle>{title}</DialogTitle>
                    <DialogDescription>{description}</DialogDescription>
                </DialogHeader>
                <DialogFooter>
                    <Button onClick={hide}>확인</Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
}

interface ConfirmDialogStore {
    isOpen: boolean;
    title: string;
    description: string;
    onConfirm: () => void;
    onCancel: () => void;
    show: (title: string, description: string, onConfirm: () => void, onCancel?: () => void) => void;
    hide: () => void;
}

export const useConfirmDialog = create<ConfirmDialogStore>((set) => ({
    isOpen: false,
    title: "",
    description: "",
    onConfirm: () => { },
    onCancel: () => { },
    show: (title, description, onConfirm, onCancel) =>
        set({ isOpen: true, title, description, onConfirm, onCancel: onCancel || (() => { }) }),
    hide: () => set({ isOpen: false }),
}));

export function ConfirmDialog() {
    const { isOpen, title, description, onConfirm, onCancel, hide } = useConfirmDialog();

    const handleConfirm = () => {
        onConfirm();
        hide();
    };

    const handleCancel = () => {
        onCancel();
        hide();
    };

    return (
        <Dialog open={isOpen} onOpenChange={hide}>
            <DialogContent>
                <DialogHeader>
                    <DialogTitle>{title}</DialogTitle>
                    <DialogDescription>{description}</DialogDescription>
                </DialogHeader>
                <DialogFooter className="flex gap-2">
                    <Button variant="outline" onClick={handleCancel}>취소</Button>
                    <Button onClick={handleConfirm}>확인</Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
}
