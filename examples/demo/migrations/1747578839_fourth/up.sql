CREATE TABLE employee_projects (
    employee_id INT,
    project_id UUID,
    assigned_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    role VARCHAR(50),
    PRIMARY KEY (employee_id, project_id),
    FOREIGN KEY (employee_id) REFERENCES employees(employee_id),
    FOREIGN KEY (project_id) REFERENCES projects(project_id) ON DELETE CASCADE
);
