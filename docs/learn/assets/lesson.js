document.documentElement.classList.add("js");

for (const quiz of document.querySelectorAll("[data-quiz]")) {
  const feedback = quiz.querySelector("[data-feedback]");
  const expected = quiz.dataset.answer;

  quiz.addEventListener("submit", (event) => {
    event.preventDefault();

    const selected = new FormData(quiz).get("answer");
    if (!selected) {
      feedback.textContent = "Choose one answer before checking.";
      feedback.dataset.state = "incorrect";
      return;
    }

    const correct = selected === expected;
    feedback.textContent = correct
      ? quiz.dataset.correctMessage
      : quiz.dataset.incorrectMessage;
    feedback.dataset.state = correct ? "correct" : "incorrect";
  });

  quiz.addEventListener("change", () => {
    feedback.textContent = "";
    delete feedback.dataset.state;
  });
}
