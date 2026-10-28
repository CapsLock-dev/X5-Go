package main

import (
	"errors"
	"fmt"
	"lesson5/student/practice"
)

func main() {
	// TODO: домашняя работа. Покажите успешную рассылку, пустой текст,
	// ошибку настройки и продолжение после BrokenNotifier.
	// Импорт учебного пакета: "lesson5/student/practice".

	email := &practice.EmailNotifier{
		BaseNotifier: practice.BaseNotifier{Channel: "email"},
		Address:      "mail@test.com",
	}
	push := &practice.PushNotifier{
		BaseNotifier: practice.BaseNotifier{Channel: "push"},
		DeviceID:      "test_device_id",
	}

	fmt.Println("\nСлучай Успех")
	practice.Dispatch([]practice.Notifier{email, push}, "заказ готов");
	fmt.Println(email.Messages);
	fmt.Println(push.Messages);

	fmt.Println("\nСлучай Пустой текст")
	err := practice.Dispatch([]practice.Notifier{email, push}, "");
	if errors.Is(err, practice.ErrEmptyMessage) {
		fmt.Println(email.Messages) 
		fmt.Println(push.Messages)
	}

	fmt.Println("\nСлучай Нет настройки")
	push.DeviceID = ""
	var send_err *practice.SendError;
	err = practice.Dispatch([]practice.Notifier{email, push}, "непустое сообщение");
	if errors.As(err, &send_err) {
		fmt.Printf("канал: %s, причина: %s\n", send_err.Channel, send_err.Reason);
	}

	fmt.Println("\nСлучай Сломанная реализация")
	broken := &practice.BrokenNotifier{
		BaseNotifier: practice.BaseNotifier{Channel: "broken_channel"},
	}
	push.DeviceID = "test_device_id"
	errs := practice.SafeDispatch([]practice.Notifier{broken, push}, "сообщение")
	for _,e := range errs {
		fmt.Println(e);
		if (errors.Is(e, practice.ErrNotifierPanic)) {
			fmt.Println("паника ErrNotifierPanic");
		}
	}
	fmt.Println(push.Messages);
}
