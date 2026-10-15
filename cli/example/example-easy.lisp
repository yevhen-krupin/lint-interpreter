(defun doublen (n) (* n 2))
(doublen 5)
(defun fib (n) (if (< n 2) n (+ (fib (- n 1)) (fib (- n 2)))))
(fib 7)
(defun fact (n) (if (<= n 1) 1 (* n (fact (- n 1)))))
(fact 5)
;; comment whatever