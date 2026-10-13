(defun hello ()
  (format t "Hello Coding Challenge World~%"))

(defun doublen (n)
  (* n 2))

(defun fib (n)
  (if (< n 2)
      n
      (+ (fib (- n 1))
         (fib (- n 2)))))

(defun fact (n)
  (if (<= n 1)
    1
    (* n (fact (- n 1)))))

(hello)

(format t "The double of 5 is ~D~%" (doublen 5))

(format t "Factorial of 5 is ~D~%" (fact 5))

(format t "The 7th number of the Fibonacci sequence is ~D~%" (fib 7))