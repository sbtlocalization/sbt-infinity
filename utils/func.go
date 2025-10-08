// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package utils

// Iteratee converts a function with one argument to a function with two arguments,
// ignoring the second argument (index). This helps when using with samber/lo library
// since it expects a function with two arguments.
func Iteratee[T, R any](f func(item T) R) func(item T, index int) R {
	return func(item T, _ int) R {
		return f(item)
	}
}
